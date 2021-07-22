package aws

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ses"
	"jaytaylor.com/html2text"

	"github.com/silinternational/wecarry-api/domain"
)

type ObjectUrl struct {
	Url        string
	Expiration time.Time
}

type awsConfig struct {
	awsAccessKeyID     string
	awsSecretAccessKey string
	awsEndpoint        string
	awsRegion          string
	awsS3Bucket        string
	awsDisableSSL      bool
	getPresignedUrl    bool
}

// presigned URL expiration
const urlLifespan = 10 * time.Minute

func getS3ConfigFromEnv() awsConfig {
	var a awsConfig
	a.awsAccessKeyID = domain.Env.AwsAccessKeyID
	a.awsSecretAccessKey = domain.Env.AwsSecretAccessKey
	a.awsEndpoint = domain.Env.AwsS3Endpoint
	a.awsRegion = domain.Env.AwsRegion
	a.awsS3Bucket = domain.Env.AwsS3Bucket
	a.awsDisableSSL = domain.Env.AwsS3DisableSSL

	if len(a.awsEndpoint) > 0 {
		// a non-empty endpoint means minIO is in use, which doesn't support the S3 object URL scheme
		a.getPresignedUrl = true
	}
	return a
}

func createS3Service(config awsConfig) (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(config.awsAccessKeyID, config.awsSecretAccessKey, ""),
		Endpoint:         aws.String(config.awsEndpoint),
		Region:           aws.String(config.awsRegion),
		DisableSSL:       aws.Bool(config.awsDisableSSL),
		S3ForcePathStyle: aws.Bool(len(config.awsEndpoint) > 0),
	})
	svc := s3.New(sess)

	return svc, err
}

func getObjectURL(config awsConfig, svc *s3.S3, key string) (ObjectUrl, error) {
	var objectUrl ObjectUrl

	if !config.getPresignedUrl {
		objectUrl.Url = fmt.Sprintf("https://%s.s3.amazonaws.com/%s", config.awsS3Bucket, url.PathEscape(key))
		objectUrl.Expiration = time.Date(9999, time.December, 31, 0, 0, 0, 0, time.UTC)
		return objectUrl, nil
	}

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(config.awsS3Bucket),
		Key:    aws.String(key),
	})

	if newUrl, err := req.Presign(urlLifespan); err == nil {
		objectUrl.Url = newUrl
		// return a time slightly before the actual url expiration to account for delays
		objectUrl.Expiration = time.Now().Add(urlLifespan - time.Minute)
	} else {
		return objectUrl, err
	}

	return objectUrl, nil
}

// StoreFile saves content in an AWS S3 bucket or compatible storage, depending on environment configuration.
func StoreFile(key, contentType string, content []byte) (ObjectUrl, error) {
	config := getS3ConfigFromEnv()

	svc, err := createS3Service(config)
	if err != nil {
		return ObjectUrl{}, err
	}

	acl := ""
	if !config.getPresignedUrl {
		acl = "public-read"
	}
	if _, err := svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(config.awsS3Bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		ACL:         aws.String(acl),
		Body:        bytes.NewReader(content),
	}); err != nil {
		return ObjectUrl{}, err
	}

	objectUrl, err := getObjectURL(config, svc, key)
	if err != nil {
		return ObjectUrl{}, err
	}

	return objectUrl, nil
}

// GetFileURL retrieves a URL from which a stored object can be loaded. The URL should not require external
// credentials to access. It may reference a file with public_read access or it may be a pre-signed URL.
func GetFileURL(key string) (ObjectUrl, error) {
	config := getS3ConfigFromEnv()

	svc, err := createS3Service(config)
	if err != nil {
		return ObjectUrl{}, err
	}

	return getObjectURL(config, svc, key)
}

// RemoveFile removes a file from the configured AWS S3 bucket.
func RemoveFile(key string) error {
	config := getS3ConfigFromEnv()

	svc, err := createS3Service(config)
	if err != nil {
		return err
	}

	if _, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(config.awsS3Bucket),
		Key:    aws.String(key),
	}); err != nil {
		return err
	}

	return nil
}

// CreateS3Bucket creates an S3 bucket with a name defined by an environment variable. If the bucket already
// exists, it will not return an error.
func CreateS3Bucket() error {
	env := domain.Env.GoEnv
	if env != "test" && env != "development" {
		return errors.New("CreateS3Bucket should only be used in test and development")
	}

	config := getS3ConfigFromEnv()

	svc, err := createS3Service(config)
	if err != nil {
		return err
	}

	c := &s3.CreateBucketInput{Bucket: &domain.Env.AwsS3Bucket}
	if _, err := svc.CreateBucket(c); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
			case s3.ErrCodeBucketAlreadyOwnedByYou:
			default:
				return err
			}
		}
	}
	return nil
}

// SendEmail sends a message using SES
func SendEmail(to, from, subject, body string) error {
	svc, err := createSESService(getSESConfigFromEnv())
	if err != nil {
		return fmt.Errorf("SendEmail failed creating SES service, %s", err)
	}

	input := &ses.SendRawEmailInput{
		RawMessage: &ses.RawMessage{Data: rawEmail(to, from, subject, body)},
		Source:     aws.String(from),
	}

	result, err := svc.SendRawEmail(input)
	if err != nil {
		return fmt.Errorf("SendEmail failed using SES, %s", err)
	}

	domain.Logger.Printf("Message sent using SES, message ID: %s", *result.MessageId)
	return nil
}

// rawEmail generates a multi-part MIME email message with a plain text, html text, and inline logo attachment as
// follows:
//
// * multipart/alternative
//   * text/plain
//   * multipart/related
//     * text/html
//     * image/png
//
// Abbreviated example of the generated email message:
//  From: from@example.com
//	To: to@example.com
//	Subject: subject text
//	Content-Type: multipart/alternative; boundary="boundary_alternative"
//
//	--boundary_alternative
//	Content-Type: text/plain; charset=utf-8
//
//	Plain text body
//	--boundary_alternative
//	Content-type: multipart/related; boundary="boundary_related"
//
//	--boundary_related
//	Content-Type: text/html; charset=utf-8
//
//	HTML body
//	--boundary_related
//	Content-Type: image/png
//	Content-Transfer-Encoding: base64
//	Content-ID: <logo>
//	--boundary_related--
//	--boundary_alternative--
func rawEmail(to, from, subject, body string) []byte {
	tbody, err := html2text.FromString(body)
	if err != nil {
		domain.Logger.Printf("error converting html email to plain text ... %s", err.Error())
		tbody = body
	}

	b := &bytes.Buffer{}

	b.WriteString("From: " + from + "\n")
	b.WriteString("To: " + to + "\n")
	b.WriteString("Subject: " + subject + "\n")
	b.WriteString("MIME-Version: 1.0\n")

	alternativeWriter := multipart.NewWriter(b)
	b.WriteString(`Content-Type: multipart/alternative; type="text/plain"; boundary="` +
		alternativeWriter.Boundary() + `"` + "\n\n")

	w, err := alternativeWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type":        {"text/plain; charset=utf-8"},
		"Content-Disposition": {"inline"},
	})
	if err != nil {
		domain.ErrLogger.Printf("failed to create MIME text part, %s", err)
	} else {
		_, _ = fmt.Fprint(w, tbody)
	}

	relatedWriter := multipart.NewWriter(b)
	_, err = alternativeWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type": {`multipart/related; type="text/html"; boundary="` + relatedWriter.Boundary() + `"`},
	})
	if err != nil {
		domain.ErrLogger.Printf("failed to create MIME related part, %s", err)
	}

	w, err = relatedWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type":        {"text/html; charset=utf-8"},
		"Content-Disposition": {"inline"},
	})
	if err != nil {
		domain.ErrLogger.Printf("failed to create MIME html part, %s", err)
	} else {
		_, _ = fmt.Fprint(w, body)
	}

	w, err = relatedWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"image/png"},
		"Content-Disposition":       {"inline"},
		"Content-ID":                {"<logo>"},
		"Content-Transfer-Encoding": {"base64"},
	})
	if err != nil {
		domain.ErrLogger.Printf("failed to create MIME image part, %s", err)
	} else {
		logo, err := domain.Assets.Find("logo.png")
		if err != nil {
			domain.ErrLogger.Printf("failed to read logo file, %s", err)
		}
		encoder := base64.NewEncoder(base64.StdEncoding, b)
		_, err = encoder.Write(logo)
		if err != nil {
			domain.ErrLogger.Printf("failed to write logo to email, %s", err)
		}
		err = encoder.Close()
		if err != nil {
			domain.ErrLogger.Printf("failed to close logo base64 encoder, %s", err)
		}
	}

	if err = relatedWriter.Close(); err != nil {
		domain.ErrLogger.Printf("failed to close MIME related part, %s", err)
	}

	if err = alternativeWriter.Close(); err != nil {
		domain.ErrLogger.Printf("failed to close MIME alternative part, %s", err)
	}

	return b.Bytes()
}

func getSESConfigFromEnv() awsConfig {
	return awsConfig{
		awsAccessKeyID:     domain.Env.AwsAccessKeyID,
		awsSecretAccessKey: domain.Env.AwsSecretAccessKey,
		awsRegion:          domain.Env.AwsRegion,
	}
}

func createSESService(config awsConfig) (*ses.SES, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(config.awsAccessKeyID, config.awsSecretAccessKey, ""),
		Region:      aws.String(config.awsRegion),
	})
	if err != nil {
		return nil, err
	}
	return ses.New(sess), nil
}
