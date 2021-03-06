package clientauthorization

import (
	"errors"
	"fmt"

	kapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	kerrors "github.com/GoogleCloudPlatform/kubernetes/pkg/api/errors"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/apiserver"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"

	"github.com/openshift/origin/pkg/oauth/api"
	"github.com/openshift/origin/pkg/oauth/api/validation"
)

// REST implements the RESTStorage interface in terms of an Registry.
type REST struct {
	registry Registry
}

// NewStorage returns a new REST.
func NewREST(registry Registry) apiserver.RESTStorage {
	return &REST{registry}
}

// New returns a new ClientAuthorization for use with Create and Update.
func (s *REST) New() runtime.Object {
	return &api.ClientAuthorization{}
}

// Get retrieves an ClientAuthorization by id.
func (s *REST) Get(ctx kapi.Context, id string) (runtime.Object, error) {
	authorization, err := s.registry.GetClientAuthorization(id)
	if err != nil {
		return nil, err
	}
	return authorization, nil
}

// List retrieves a list of ClientAuthorizations that match selector.
func (s *REST) List(ctx kapi.Context, label, fields labels.Selector) (runtime.Object, error) {
	return s.registry.ListClientAuthorizations(label, labels.Everything())
}

// Create registers the given ClientAuthorization.
func (s *REST) Create(ctx kapi.Context, obj runtime.Object) (<-chan apiserver.RESTResult, error) {
	authorization, ok := obj.(*api.ClientAuthorization)
	if !ok {
		return nil, fmt.Errorf("not an authorization: %#v", obj)
	}

	if authorization.UserName == "" || authorization.ClientName == "" {
		return nil, errors.New("invalid authorization")
	}

	authorization.Name = s.registry.ClientAuthorizationName(authorization.UserName, authorization.ClientName)
	kapi.FillObjectMetaSystemFields(ctx, &authorization.ObjectMeta)

	if errs := validation.ValidateClientAuthorization(authorization); len(errs) > 0 {
		return nil, kerrors.NewInvalid("clientAuthorization", authorization.Name, errs)
	}

	return apiserver.MakeAsync(func() (runtime.Object, error) {
		if err := s.registry.CreateClientAuthorization(authorization); err != nil {
			return nil, err
		}
		return s.Get(ctx, authorization.Name)
	}), nil
}

// Update modifies an existing client authorization
func (s *REST) Update(ctx kapi.Context, obj runtime.Object) (<-chan apiserver.RESTResult, error) {
	authorization, ok := obj.(*api.ClientAuthorization)
	if !ok {
		return nil, fmt.Errorf("not an authorization: %#v", obj)
	}

	if errs := validation.ValidateClientAuthorization(authorization); len(errs) > 0 {
		return nil, kerrors.NewInvalid("clientAuthorization", authorization.Name, errs)
	}

	oldauth, err := s.registry.GetClientAuthorization(authorization.Name)
	if err != nil {
		return nil, err
	}
	if errs := validation.ValidateClientAuthorizationUpdate(authorization, oldauth); len(errs) > 0 {
		return nil, kerrors.NewInvalid("clientAuthorization", authorization.Name, errs)
	}

	return apiserver.MakeAsync(func() (runtime.Object, error) {
		if err := s.registry.UpdateClientAuthorization(authorization); err != nil {
			return nil, err
		}
		return s.Get(ctx, authorization.Name)
	}), nil
}

// Delete asynchronously deletes an ClientAuthorization specified by its id.
func (s *REST) Delete(ctx kapi.Context, id string) (<-chan apiserver.RESTResult, error) {
	return apiserver.MakeAsync(func() (runtime.Object, error) {
		return &kapi.Status{Status: kapi.StatusSuccess}, s.registry.DeleteClientAuthorization(id)
	}), nil
}
