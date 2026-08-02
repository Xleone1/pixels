package condition

import (
	"strings"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
)

// Evaluate evaluates one canonical condition and fails closed on dependency errors.
func (evaluator Evaluator) Evaluate(node *configuration.Node, context Context, view View) (Result, error) {
	if node == nil || view == nil {
		return Result{}, nil
	}
	key, negative := positiveKey(node.Descriptor.Key)
	if strings.HasPrefix(key, "plugin.") && evaluator.extensions != nil {
		result, found, err := evaluator.extensions.EvaluateCondition(node, context)
		if found {
			return result, err
		}
	}
	result, err := evaluatePositive(key, node, context, view)
	if err != nil || !result.Valid {
		return Result{Valid: result.Valid}, err
	}
	if negative {
		result.Pass = !result.Pass
	}
	return result, nil
}

// evaluatePositive evaluates one positive predicate.
func evaluatePositive(key string, node *configuration.Node, context Context, view View) (Result, error) {
	switch key {
	case "wf_cnd_match_snapshot":
		return everyTarget(node, func(itemID int64, index int) (bool, error) {
			return view.SnapshotMatches(itemID, node.Targets[index].Snapshot, node.Parameters.Values)
		})
	case "wf_cnd_furnis_hv_avtrs":
		return everyTarget(node, func(itemID int64, _ int) (bool, error) { return view.UnitOn(itemID) })
	case "wf_cnd_trggrer_on_frn":
		return actorOnTargets(node, context, view)
	case "wf_cnd_time_more_than":
		return Result{Pass: context.Now.Sub(context.ResetAt) > node.Parameters.Duration, Valid: !context.ResetAt.IsZero()}, nil
	case "wf_cnd_time_less_than":
		return Result{Pass: context.Now.Sub(context.ResetAt) < node.Parameters.Duration, Valid: !context.ResetAt.IsZero()}, nil
	case "wf_cnd_user_count_in":
		minimum, maximum := int(node.Parameters.Values[0]), int(node.Parameters.Values[1])
		count := view.UserCount()
		return Result{Pass: count >= minimum && count <= maximum, Valid: true}, nil
	case "wf_cnd_actor_in_team":
		pass, valid, err := view.ActorTeam(context.Event.PlayerID, node.Parameters.Values[0])
		return Result{Pass: pass, Valid: valid}, err
	case "wf_cnd_has_furni_on":
		return stackedTargets(node, view)
	case "wf_cnd_stuff_is":
		return Result{Pass: targetMatches(node, context.Event.SourceItem, context.Event.SourceSprite), Valid: context.Event.SourceItem > 0}, nil
	case "wf_cnd_actor_in_group":
		pass, valid, err := view.ActorGroup(context.Event.PlayerID)
		return Result{Pass: pass, Valid: valid}, err
	case "wf_cnd_wearing_badge":
		pass, valid, err := view.WearingBadge(context.Event.PlayerID, strings.TrimSpace(node.Parameters.Text))
		return Result{Pass: pass, Valid: valid}, err
	case "wf_cnd_wearing_effect":
		pass, valid, err := view.WearingEffect(context.Event.PlayerID, first(node.Parameters.Values))
		return Result{Pass: pass, Valid: valid}, err
	case "wf_cnd_date_rng_active":
		start := time.Unix(int64(node.Parameters.Values[0]), 0)
		end := time.Unix(int64(node.Parameters.Values[1]), 0)
		return Result{Pass: !context.Now.Before(start) && !context.Now.After(end), Valid: true}, nil
	case "wf_cnd_has_handitem":
		pass, valid, err := view.HasHanditem(context.Event.PlayerID, first(node.Parameters.Values))
		return Result{Pass: pass, Valid: valid}, err
	case "wf_cnd_user_performs_action":
		pass, valid, err := view.PerformsAction(context.Event, node.Parameters.Values)
		return Result{Pass: pass, Valid: valid}, err
	case "wf_cnd_team_has_rank":
		pass, valid, err := view.TeamHasRank(first(node.Parameters.Values), value(node.Parameters.Values, 1))
		return Result{Pass: pass, Valid: valid}, err
	case "wf_cnd_has_altitude":
		return altitudeMatches(node, view)
	case "wf_cnd_actor_dir":
		return Result{Pass: context.Event.ActorID > 0 && context.Event.Direction == first(node.Parameters.Values), Valid: context.Event.ActorID > 0}, nil
	case "wf_cnd_slc_quantity":
		count := int64(len(context.Selection.Furni) + len(context.Selection.Actors))
		return Result{Pass: compare(count, int64(value(node.Parameters.Values, 1)), first(node.Parameters.Values)), Valid: true}, nil
	case "wf_cnd_clock_matches":
		return Result{Pass: compare(view.ClockCounter(), int64(value(node.Parameters.Values, 1)), first(node.Parameters.Values)), Valid: true}, nil
	case "wf_cnd_has_var":
		return variableExists(node, context)
	case "wf_cnd_var_val_match":
		return variableValueMatches(node, context, false)
	case "wf_cnd_var_age_match":
		return variableValueMatches(node, context, true)
	case "wf_cnd_valid_moves":
		pass, err := view.ValidMoves(context.Effects, context.Event)
		return Result{Pass: pass, Valid: true}, err
	default:
		return Result{}, nil
	}
}

// positiveKey maps negative descriptors to their positive predicate.
func positiveKey(key string) (string, bool) {
	switch key {
	case "wf_cnd_not_match_snap":
		return "wf_cnd_match_snapshot", true
	case "wf_cnd_not_hv_avtrs":
		return "wf_cnd_furnis_hv_avtrs", true
	case "wf_cnd_not_trggrer_on":
		return "wf_cnd_trggrer_on_frn", true
	case "wf_cnd_not_user_count":
		return "wf_cnd_user_count_in", true
	case "wf_cnd_not_in_team":
		return "wf_cnd_actor_in_team", true
	case "wf_cnd_not_furni_on":
		return "wf_cnd_has_furni_on", true
	case "wf_cnd_not_stuff_is":
		return "wf_cnd_stuff_is", true
	case "wf_cnd_not_in_group":
		return "wf_cnd_actor_in_group", true
	case "wf_cnd_not_wearing_b":
		return "wf_cnd_wearing_badge", true
	case "wf_cnd_not_wearing_fx":
		return "wf_cnd_wearing_effect", true
	case "wf_cnd_not_has_altitude":
		return "wf_cnd_has_altitude", true
	case "wf_cnd_neg_has_var":
		return "wf_cnd_has_var", true
	case "wf_cnd_not_has_handitem":
		return "wf_cnd_has_handitem", true
	case "wf_cnd_not_user_performs_action":
		return "wf_cnd_user_performs_action", true
	default:
		return key, false
	}
}
