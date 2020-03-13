package clients

import (
	"errors"
	"github.com/herbal828/ci_cd-api/api/utils"
	"reflect"
	"testing"

	"github.com/mercadolibre/golang-restclient/rest"
)

func Test_client_Get(t *testing.T) {
	type fields struct {
		RestClient *rest.RequestBuilder
	}
	type args struct {
		url string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   Response
	}{
		{
			name: "test just running the Get func",
			args: args{
				url: "url_test",
			},
			fields: fields{
				RestClient: &rest.RequestBuilder{
					BaseURL: "http://testbaseurl.com",
				},
			},
			want: newResponse(&rest.Response{
				Err: errors.New("algo salio mal"),
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				RestClient: tt.fields.RestClient,
			}
			if got := c.Get(tt.args.url); reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_client_Post(t *testing.T) {
	type fields struct {
		RestClient *rest.RequestBuilder
	}
	type args struct {
		url  string
		body interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Response
		body   interface{}
	}{
		{
			name: "test just running the Post func",
			args: args{
				url: "url_test",
			},
			fields: fields{
				RestClient: &rest.RequestBuilder{
					BaseURL: "http://testbaseurl.com",
				},
			},
			want: newResponse(&rest.Response{
				Err: errors.New("algo salio mal"),
			}),
			body: map[string]interface{}{
				"repository_name": "fury_repo-name",
				"type":            "gitflow",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				RestClient: tt.fields.RestClient,
			}
			if got := c.Post(tt.args.url, tt.args.body); reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.Post() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_client_Put(t *testing.T) {
	type fields struct {
		Client     Client
		RestClient *rest.RequestBuilder
	}
	type args struct {
		url  string
		body interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Response
		body   interface{}
	}{
		{
			name: "test just running the Put func",
			args: args{
				url: "url_test",
			},
			fields: fields{
				RestClient: &rest.RequestBuilder{
					BaseURL: "http://testbaseurl.com",
				},
			},
			want: newResponse(&rest.Response{
				Err: errors.New("algo salio mal"),
			}),
			body: map[string]interface{}{
				"repository_name": "fury_repo-name",
				"type":            "gitflow",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				RestClient: tt.fields.RestClient,
			}
			if got := c.Put(tt.args.url, tt.args.body); reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.Put() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_client_Delete(t *testing.T) {
	type fields struct {
		RestClient *rest.RequestBuilder
	}
	type args struct {
		url string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Response
	}{
		{
			name: "test just running the Delete func",
			args: args{
				url: "url_test",
			},
			fields: fields{
				RestClient: &rest.RequestBuilder{
					BaseURL: "http://testbaseurl.com",
				},
			},
			want: newResponse(&rest.Response{
				Err: errors.New("algo salio mal"),
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				RestClient: tt.fields.RestClient,
			}
			if got := c.Delete(tt.args.url); reflect.DeepEqual(got, tt.want) {
				t.Errorf("client.Delete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_response_Bytes(t *testing.T) {
	type fields struct {
		Response *rest.Response
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "test just running the Bytes func",
			fields: fields{
				Response: &rest.Response{},
			},
			want: utils.GetBytes(nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &response{
				Response: tt.fields.Response,
			}
			if got := r.Bytes(); reflect.DeepEqual(got, tt.want) {
				t.Errorf("response.Bytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_response_Err(t *testing.T) {
	type fields struct {
		Response *rest.Response
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test just running the Err func",
			fields: fields{
				Response: &rest.Response{
					Err: errors.New("some error"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &response{
				Response: tt.fields.Response,
			}
			if err := r.Err(); (err != nil) != tt.wantErr {
				t.Errorf("response.Err() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}