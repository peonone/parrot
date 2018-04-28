package srv

import (
	"github.com/micro/go-micro/registry"
	"github.com/stretchr/testify/mock"
)

type mockStateStore struct {
	mock.Mock
}

func (s *mockStateStore) online(uid string, webNode string) error {
	returnVals := s.Called(uid, webNode)
	return returnVals.Error(0)
}

func (s *mockStateStore) offline(uid string) error {
	returnVals := s.Called(uid)
	return returnVals.Error(0)
}

func (s *mockStateStore) getOnlineWebNode(uid string) (string, error) {
	returnValues := s.Called(uid)
	return returnValues.String(0), returnValues.Error(1)
}

type mockMQSender struct {
	mock.Mock
}

func (s *mockMQSender) sendMQMsg(key string, body []byte) error {
	returnVals := s.Called(key, body)
	return returnVals.Error(0)
}

type mockRegistry struct {
	mock.Mock
}

func (r *mockRegistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	args := make([]interface{}, 0, len(opts)+1)
	args = append(args, s)
	for _, opt := range opts {
		args = append(args, opt)
	}
	return r.Called(args...).Error(0)
}

func (r *mockRegistry) Deregister(s *registry.Service) error {
	return r.Called(s).Error(0)
}

func (r *mockRegistry) GetService(srvName string) ([]*registry.Service, error) {
	returnValues := r.Called(srvName)
	return returnValues.Get(0).([]*registry.Service), returnValues.Error(1)
}

func (r *mockRegistry) ListServices() ([]*registry.Service, error) {
	returnValues := r.Called()
	return returnValues.Get(0).([]*registry.Service), returnValues.Error(1)
}

func (r *mockRegistry) Watch(opts ...registry.WatchOption) (registry.Watcher, error) {
	args := make([]interface{}, 0, len(opts))
	for _, opt := range opts {
		args = append(args, opt)
	}
	returnVals := r.Called(args...)
	return returnVals.Get(0).(registry.Watcher), returnVals.Error(1)
}

func (r *mockRegistry) String() string {
	return r.Called().String(0)
}

func (r *mockRegistry) Options() registry.Options {
	return r.Called().Get(0).(registry.Options)
}
