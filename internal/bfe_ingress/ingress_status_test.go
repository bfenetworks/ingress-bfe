package bfe_ingress

import (
	"testing"

	"github.com/bfenetworks/ingress-bfe/internal/kubernetes_client"
)

func TestIngressStatusWriter_getErrorMsg(t *testing.T) {
	type fields struct {
		client *kubernetes_client.KubernetesClient
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
		{
			name:   "normal",
			fields: fields{client: nil},
			args:   args{msg: "error msg"},
			want:   "{\"status\":\"error\",\"message\":\"error msg\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &IngressStatusWriter{
				client: tt.fields.client,
			}
			if got := w.getErrorMsg(tt.args.msg); got != tt.want {
				t.Errorf("IngressStatusWriter.getErrorMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIngressStatusWriter_getSuccessMsg(t *testing.T) {
	type fields struct {
		client *kubernetes_client.KubernetesClient
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
		{
			name:   "normal",
			fields: fields{client: nil},
			args:   args{msg: "success msg"},
			want:   "{\"status\":\"success\",\"message\":\"success msg\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &IngressStatusWriter{
				client: tt.fields.client,
			}
			if got := w.getSuccessMsg(tt.args.msg); got != tt.want {
				t.Errorf("IngressStatusWriter.getSuccessMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}
