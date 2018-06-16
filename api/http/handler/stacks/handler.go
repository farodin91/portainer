package stacks

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/portainer/portainer"
	httperror "github.com/portainer/portainer/http/error"
	"github.com/portainer/portainer/http/security"
)

// Handler is the HTTP handler used to handle stack operations.
type Handler struct {
	stackCreationMutex *sync.Mutex
	stackDeletionMutex *sync.Mutex
	*mux.Router
	FileService            portainer.FileService
	GitService             portainer.GitService
	StackService           portainer.StackService
	EndpointService        portainer.EndpointService
	EndpointGroupService   portainer.EndpointGroupService
	TeamMembershipService  portainer.TeamMembershipService
	ResourceControlService portainer.ResourceControlService
	RegistryService        portainer.RegistryService
	DockerHubService       portainer.DockerHubService
	SwarmStackManager      portainer.SwarmStackManager
	ComposeStackManager    portainer.ComposeStackManager
}

// NewHandler creates a handler to manage stack operations.
func NewHandler(bouncer *security.RequestBouncer) *Handler {
	h := &Handler{
		Router:             mux.NewRouter(),
		stackCreationMutex: &sync.Mutex{},
		stackDeletionMutex: &sync.Mutex{},
	}
	h.Handle("/stacks",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.stackCreate))).Methods(http.MethodPost)
	h.Handle("/stacks",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.stackList))).Methods(http.MethodGet)
	h.Handle("/stacks/{id}",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.stackInspect))).Methods(http.MethodGet)
	h.Handle("/stacks/{id}",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.stackDelete))).Methods(http.MethodDelete)
	h.Handle("/stacks/{id}",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.stackUpdate))).Methods(http.MethodPut)
	h.Handle("/stacks/{id}/file",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.stackFile))).Methods(http.MethodGet)
	h.Handle("/stacks/{id}/migrate",
		bouncer.RestrictedAccess(httperror.LoggerHandler(h.stackMigrate))).Methods(http.MethodPost)
	return h
}

func (handler *Handler) checkEndpointAccess(endpoint *portainer.Endpoint, userID portainer.UserID) error {
	memberships, err := handler.TeamMembershipService.TeamMembershipsByUserID(userID)
	if err != nil {
		return err
	}

	group, err := handler.EndpointGroupService.EndpointGroup(endpoint.GroupID)
	if err != nil {
		return err
	}

	if !security.AuthorizedEndpointAccess(endpoint, group, userID, memberships) {
		return portainer.ErrEndpointAccessDenied
	}

	return nil
}
