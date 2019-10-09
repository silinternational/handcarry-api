package domain

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/gofrs/uuid"
)

func TestGetFirstStringFromSlice(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nil",
			args: args{
				s: nil,
			},
			want: "",
		},
		{
			name: "empty slice",
			args: args{
				s: []string{},
			},
			want: "",
		},
		{
			name: "single string in slice",
			args: args{
				s: []string{"alpha"},
			},
			want: "alpha",
		},
		{
			name: "two strings in slice",
			args: args{
				s: []string{"alpha", "beta"},
			},
			want: "alpha",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFirstStringFromSlice(tt.args.s); got != tt.want {
				t.Errorf("GetFirstStringFromSlice() = \"%v\", want \"%v\"", got, tt.want)
			}
		})
	}
}

func TestGetBearerTokenFromRequest(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid",
			args: args{
				r: &http.Request{
					Header: map[string][]string{
						"Authorization": {"Bearer abc123"},
					},
				},
			},
			want: "abc123",
		},
		{
			name: "also valid, not case-sensitive",
			args: args{
				r: &http.Request{
					Header: map[string][]string{
						"Authorization": {"bearer def456"},
					},
				},
			},
			want: "def456",
		},
		{
			name: "missing authorization header",
			args: args{
				r: &http.Request{
					Header: map[string][]string{
						"Other": {"Bearer abc123"},
					},
				},
			},
			want: "",
		},
		{
			name: "valid, but more complicated",
			args: args{
				r: &http.Request{
					Header: map[string][]string{
						"Authorization": {"Bearer 861B1C06-DDB8-494F-8627-3A87B22FFB82"},
					},
				},
			},
			want: "861B1C06-DDB8-494F-8627-3A87B22FFB82",
		},
		{
			name: "invalid format, missing bearer",
			args: args{
				r: &http.Request{
					Header: map[string][]string{
						"Authorization": {"861B1C06-DDB8-494F-8627-3A87B22FFB82"},
					},
				},
			},
			want: "",
		},
		{
			name: "invalid format, has colon",
			args: args{
				r: &http.Request{
					Header: map[string][]string{
						"Authorization": {"Bearer: 861B1C06-DDB8-494F-8627-3A87B22FFB82"},
					},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBearerTokenFromRequest(tt.args.r); got != tt.want {
				t.Errorf("GetBearerTokenFromRequest() = \"%v\", want \"%v\"", got, tt.want)
			}
		})
	}
}

func TestGetSubPartKeyValues(t *testing.T) {
	type args struct {
		inString, outerDelimiter, innerDelimiter string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "empty string",
			args: args{
				inString:       "",
				outerDelimiter: "!",
				innerDelimiter: "*",
			},
			want: map[string]string{},
		},
		{
			name: "one pair",
			args: args{
				inString:       "param^value",
				outerDelimiter: "#",
				innerDelimiter: "^",
			},
			want: map[string]string{
				"param": "value",
			},
		},
		{
			name: "two pairs",
			args: args{
				inString:       "param1(value1@param2(value2",
				outerDelimiter: "@",
				innerDelimiter: "(",
			},
			want: map[string]string{
				"param1": "value1",
				"param2": "value2",
			},
		},
		{
			name: "no inner delimiter",
			args: args{
				inString:       "param-value",
				outerDelimiter: "-",
				innerDelimiter: "=",
			},
			want: map[string]string{},
		},
		{
			name: "extra inner delimiter",
			args: args{
				inString:       "param=value=extra",
				outerDelimiter: "-",
				innerDelimiter: "=",
			},
			want: map[string]string{},
		},
		{
			name: "empty value",
			args: args{
				inString:       "param=value-empty=",
				outerDelimiter: "-",
				innerDelimiter: "=",
			},
			want: map[string]string{
				"param": "value",
				"empty": "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSubPartKeyValues(tt.args.inString, tt.args.outerDelimiter, tt.args.innerDelimiter)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSubPartKeyValues() = \"%v\", want \"%v\"", got, tt.want)
			}
		})
	}
}

func TestConvertTimeToStringPtr(t *testing.T) {
	now := time.Now()
	type args struct {
		inTime time.Time
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default",
			args: args{
				inTime: time.Time{},
			},
			want: "0001-01-01T00:00:00Z",
		},
		{
			name: "now",
			args: args{
				inTime: now,
			},
			want: now.Format(time.RFC3339),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ConvertTimeToStringPtr(test.args.inTime)
			if *got != test.want {
				t.Errorf("ConvertTimeToStringPtr() = \"%v\", want \"%v\"", *got, test.want)
			}
		})
	}
}

func TestConvertStringPtrToDate(t *testing.T) {
	testTime := time.Date(2019, time.August, 12, 0, 0, 0, 0, time.UTC)
	testStr := testTime.Format("2006-01-02") // not using a const in order to detect code changes
	emptyStr := ""
	badTime := "1"
	type args struct {
		inPtr *string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{{
		name: "nil",
		args: args{nil},
		want: time.Time{},
	}, {
		name: "empty",
		args: args{&emptyStr},
		want: time.Time{},
	}, {
		name: "good",
		args: args{&testStr},
		want: testTime,
	}, {
		name:    "error",
		args:    args{&badTime},
		wantErr: true,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ConvertStringPtrToDate(test.args.inPtr)
			if test.wantErr == false && err != nil {
				t.Errorf("Unexpected error %v", err)
			} else if got != test.want {
				t.Errorf("ConvertStringPtrToDate() = \"%v\", want \"%v\"", got, test.want)
			}
		})
	}
}

func TestIsStringInSlice(t *testing.T) {
	type TestData struct {
		Needle   string
		Haystack []string
		Expected bool
	}

	allTestData := []TestData{
		{
			Needle:   "no",
			Haystack: []string{},
			Expected: false,
		},
		{
			Needle:   "no",
			Haystack: []string{"really", "are you sure"},
			Expected: false,
		},
		{
			Needle:   "yes",
			Haystack: []string{"yes"},
			Expected: true,
		},
		{
			Needle:   "yes",
			Haystack: []string{"one", "two", "three", "yes"},
			Expected: true,
		},
	}

	for i, td := range allTestData {
		results := IsStringInSlice(td.Needle, td.Haystack)
		expected := td.Expected

		if results != expected {
			t.Errorf("Bad results for test set i = %v. Expected %v, but got %v", i, expected, results)
			return
		}
	}
}

func Test_emptyUuidValue(t *testing.T) {
	val := uuid.UUID{}
	if val.String() != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("empty uuid value not as expected, got: %s", val.String())
	}
}

func TestEmailDomain(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{name: "empty string", email: "", want: ""},
		{name: "domain only", email: "example.org", want: "example.org"},
		{name: "full email", email: "user@example.org", want: "example.org"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := EmailDomain(test.email); got != test.want {
				t.Errorf("incorrect response from EmailDomain(): %v, expected %v", got, test.want)
			}
		})
	}
}
