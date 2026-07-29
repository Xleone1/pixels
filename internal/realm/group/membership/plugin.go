package membership

import (
	"context"
	"errors"

	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouppolicy "github.com/niflaot/pixels/internal/realm/group/policy"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

var (
	// ErrCancelledByPlugin reports a membership mutation vetoed before persistence.
	ErrCancelledByPlugin = errors.New("group membership mutation cancelled by plugin")
)

// EventDispatcher intercepts group membership changes before persistence.
type EventDispatcher interface {
	// DispatchGroupMembershipChange reports whether the mutation was cancelled.
	DispatchGroupMembershipChange(context.Context, string, int64, int64, string) bool
}

// SetPluginRuntime installs the optional membership interceptor.
func (service *Service) SetPluginRuntime(events EventDispatcher) { service.pluginEvents = events }

// roleName returns the stable plugin-facing role name.
func roleName(role grouprecord.Role) string {
	switch role {
	case grouprecord.Owner:
		return "owner"
	case grouprecord.Admin:
		return "admin"
	default:
		return "member"
	}
}

// Join joins an open group or creates an exclusive-group request.
func (service *Service) Join(ctx context.Context, playerID int64, groupID int64) (grouprecord.Membership, bool, error) {
	if service.pluginEvents != nil && service.pluginEvents.DispatchGroupMembershipChange(ctx, "join", playerID, groupID, "") {
		return grouprecord.Membership{}, false, ErrCancelledByPlugin
	}
	member, pending, changed, err := service.store.Join(ctx, groupID, playerID, service.config.MembershipLimit, service.config.MemberLimit, service.config.PendingLimit)
	if err == nil && changed {
		service.projectChange(ctx, groupID, playerID, "join")
	}
	service.record(groupobservability.KindJoin, err)
	return member, pending, err
}

// Add administratively inserts one member or admin idempotently.
func (service *Service) Add(ctx context.Context, actorID int64, groupID int64, playerID int64, role grouprecord.Role) (grouprecord.Membership, bool, error) {
	if err := service.requireRosterManager(ctx, actorID, groupID); err != nil {
		return grouprecord.Membership{}, false, err
	}
	if role == grouprecord.Admin {
		allowed, err := service.has(ctx, actorID, grouppolicy.RolesManageAny)
		actor, found, memberErr := service.store.Membership(ctx, groupID, actorID)
		if memberErr != nil {
			return grouprecord.Membership{}, false, memberErr
		}
		if err != nil || !allowed && (!found || actor.Role != grouprecord.Owner) {
			return grouprecord.Membership{}, false, grouprecord.ErrForbidden
		}
	}
	if service.pluginEvents != nil && service.pluginEvents.DispatchGroupMembershipChange(ctx, "add", playerID, groupID, roleName(role)) {
		return grouprecord.Membership{}, false, ErrCancelledByPlugin
	}
	member, created, err := service.store.AddMember(ctx, groupID, playerID, role, service.config.MembershipLimit, service.config.MemberLimit)
	if err == nil {
		service.projectChange(ctx, groupID, playerID, "add")
	}
	service.record(groupobservability.KindJoin, err)
	return member, created, err
}

// Accept accepts one pending request after social-role authorization.
func (service *Service) Accept(ctx context.Context, actorID int64, groupID int64, playerID int64) (grouprecord.Membership, error) {
	if err := service.requireRosterManager(ctx, actorID, groupID); err != nil {
		return grouprecord.Membership{}, err
	}
	if service.pluginEvents != nil && service.pluginEvents.DispatchGroupMembershipChange(ctx, "accept", playerID, groupID, "") {
		return grouprecord.Membership{}, ErrCancelledByPlugin
	}
	member, err := service.store.AcceptRequest(ctx, groupID, playerID, service.config.MemberLimit)
	if err == nil {
		service.projectChange(ctx, groupID, playerID, "accept")
	}
	service.record(groupobservability.KindAccept, err)
	return member, err
}
