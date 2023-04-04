/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package echo

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

// MeshPod represents a connection to a specific pod running in the mesh.
// This can be used to trigger requests *from* that pod.
type MeshPod struct {
	Name      string
	Namespace string
	rc        *rest.RESTClient
	rcfg      *rest.Config
}

type MeshApplication string

const (
	MeshAppEchoV1 MeshApplication = "app=echo,version=v1"
	MeshAppEchoV2 MeshApplication = "app=echo,version=v2"
)

func (m *MeshPod) SendRequest(t *testing.T, exp http.ExpectedResponse) {
	r := exp.Request
	protocol := strings.ToLower(r.Protocol)
	if protocol == "" {
		protocol = "http"
	}
	args := []string{"client", fmt.Sprintf("%s://%s/%s", protocol, r.Host, r.Path)}
	if r.Method != "" {
		args = append(args, "--method="+r.Method)
	}
	if !r.UnfollowRedirect {
		args = append(args, "--follow-redirects")
	}
	for k, v := range r.Headers {
		args = append(args, "-H", fmt.Sprintf("%v: %v", k, v))
	}

	resp, err := m.request(args)
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	t.Logf("Got resp %v", resp)
	want := exp.Response
	if fmt.Sprint(want.StatusCode) != resp.Code {
		t.Errorf("wanted status code %v, got %v", want.StatusCode, resp.Code)
	}
	for _, name := range want.AbsentHeaders {
		if v := resp.ResponseHeaders.Values(name); len(v) != 0 {
			t.Errorf("expected no header %q, got %v", name, v)
		}
	}
	for k, v := range want.Headers {
		if got := resp.RequestHeaders.Get(k); got != v {
			t.Errorf("expected header %v=%v, got %v", k, v, got)
		}
	}
}

func (m *MeshPod) request(args []string) (Response, error) {
	container := "echo"

	req := m.rc.Post().
		Resource("pods").
		Name(m.Name).
		Namespace(m.Namespace).
		SubResource("exec").
		Param("container", container).
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   args,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(m.rcfg, "POST", req.URL())
	if err != nil {
		return Response{}, err
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
		Tty:    false,
	})
	if err != nil {
		return Response{}, err
	}

	return ParseResponse(stdoutBuf.String()), nil
}

func ConnectToApp(t *testing.T, s *suite.ConformanceTestSuite, app MeshApplication) MeshPod {
	// hardcoded, for now
	ns := "gateway-conformance-mesh"
	metaList := &metav1.PartialObjectMetadataList{}
	metaList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "PodList",
	})

	err := s.Client.List(context.Background(), metaList, client.InNamespace(ns), client.HasLabels(strings.Split(string(app), ",")))
	if err != nil {
		t.Fatalf("failed to query pods in app %v", app)
	}
	if len(metaList.Items) == 0 {
		t.Fatalf("no pods found in app %v", app)
	}
	podName := metaList.Items[0].Name
	podNamespace := metaList.Items[0].Namespace

	return MeshPod{
		Name:      podName,
		Namespace: podNamespace,
		rc:        s.RESTClient,
		rcfg:      s.RestConfig,
	}
}
