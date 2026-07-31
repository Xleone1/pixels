package configuration

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

// validate enforces shared and descriptor-specific schema rules.
func (compiler *Compiler) validate(stored record.Config, descriptor registry.Descriptor) error {
	if stored.ItemID <= 0 || stored.RoomID <= 0 || stored.DelayPulses < 0 || stored.DelayPulses > compiler.config.MaxDelayPulses {
		return ErrInvalid
	}
	if stored.SelectionMode < 0 || stored.SelectionMode > 3 || len(stored.Targets) > compiler.config.MaxSelection || len(stored.IntParams) > 32 || len(stored.StringParam) > 2048 {
		return ErrInvalid
	}
	if descriptor.Selection == registry.SelectionNone && len(stored.Targets) != 0 {
		return fmt.Errorf("%w: targets rejected", ErrInvalid)
	}
	if descriptor.Selection == registry.SelectionNone && stored.SelectionMode != 0 {
		return fmt.Errorf("%w: selection mode rejected", ErrInvalid)
	}
	if descriptor.Selection == registry.SelectionRequired && len(stored.Targets) == 0 {
		return fmt.Errorf("%w: targets required", ErrInvalid)
	}
	if len(stored.Targets) > 0 && stored.SelectionMode == 0 {
		return fmt.Errorf("%w: selection mode required", ErrInvalid)
	}
	seen := make(map[int64]struct{}, len(stored.Targets))
	for _, target := range stored.Targets {
		if target.ItemID <= 0 {
			return fmt.Errorf("%w: invalid target", ErrInvalid)
		}
		if _, exists := seen[target.ItemID]; exists {
			return fmt.Errorf("%w: duplicate target", ErrInvalid)
		}
		seen[target.ItemID] = struct{}{}
	}
	return validateBehavior(stored, descriptor.Key)
}

// validateBehavior enforces schemas whose editor fields have strict meaning.
func validateBehavior(stored record.Config, key string) error {
	switch key {
	case "wf_trg_says_something":
		if strings.TrimSpace(stored.StringParam) == "" || len(stored.StringParam) > 100 {
			return fmt.Errorf("%w: keyword", ErrInvalid)
		}
	case "wf_trg_periodically", "wf_trg_period_long", "wf_trg_at_given_time", "wf_trg_at_time_long", "wf_cnd_time_more_than", "wf_cnd_time_less_than":
		if !positiveAt(stored.IntParams, 0) {
			return fmt.Errorf("%w: duration", ErrInvalid)
		}
	case "wf_cnd_user_count_in", "wf_cnd_not_user_count":
		if len(stored.IntParams) < 2 || stored.IntParams[0] < 0 || stored.IntParams[1] < stored.IntParams[0] {
			return fmt.Errorf("%w: user range", ErrInvalid)
		}
	case "wf_act_join_team", "wf_act_give_score_tm", "wf_cnd_actor_in_team", "wf_cnd_not_in_team":
		if len(stored.IntParams) == 0 || stored.IntParams[0] < 1 || stored.IntParams[0] > 4 {
			return fmt.Errorf("%w: team", ErrInvalid)
		}
	case "wf_cnd_date_rng_active":
		if len(stored.IntParams) < 2 || stored.IntParams[1] < stored.IntParams[0] {
			return fmt.Errorf("%w: date range", ErrInvalid)
		}
	case "wf_act_mute_triggerer":
		if !positiveAt(stored.IntParams, 0) || stored.IntParams[0] > 1440 {
			return fmt.Errorf("%w: mute duration", ErrInvalid)
		}
	case "wf_act_give_score":
		if len(stored.IntParams) < 2 || stored.IntParams[0] < 1 || stored.IntParams[0] > 100 || stored.IntParams[1] < 1 || stored.IntParams[1] > 10 {
			return fmt.Errorf("%w: score and use limit", ErrInvalid)
		}
	case "wf_act_progress_achievement":
		if strings.TrimSpace(stored.StringParam) == "" {
			return fmt.Errorf("%w: achievement group", ErrInvalid)
		}
	case "wf_act_progress_quest", "wf_act_start_quest":
		value, err := strconv.ParseInt(strings.TrimSpace(stored.StringParam), 10, 64)
		if err != nil || value <= 0 {
			return fmt.Errorf("%w: quest id", ErrInvalid)
		}
	case "wf_slc_users_area", "wf_slc_users_neighborhood", "wf_slc_furni_area", "wf_slc_furni_neighborhood":
		if len(stored.IntParams) < 4 || stored.IntParams[0] < 0 || stored.IntParams[1] < 0 || stored.IntParams[2] < 0 || stored.IntParams[3] < 0 {
			return fmt.Errorf("%w: selector area", ErrInvalid)
		}
	case "wf_slc_users_team":
		if len(stored.IntParams) == 0 || stored.IntParams[0] < 1 || stored.IntParams[0] > 4 {
			return fmt.Errorf("%w: selector team", ErrInvalid)
		}
	case "wf_slc_users_bytype":
		if len(stored.IntParams) == 0 || stored.IntParams[0] < 1 || stored.IntParams[0] > 3 {
			return fmt.Errorf("%w: selector actor type", ErrInvalid)
		}
	case "wf_slc_furni_altitude", "wf_cnd_has_altitude", "wf_cnd_not_has_altitude":
		if len(stored.IntParams) < 2 || stored.IntParams[0] < 0 || stored.IntParams[1] < 0 {
			return fmt.Errorf("%w: altitude range", ErrInvalid)
		}
	case "wf_act_set_altitude":
		if len(stored.IntParams) == 0 || stored.IntParams[0] < 0 || stored.IntParams[0] > 10000 {
			return fmt.Errorf("%w: altitude", ErrInvalid)
		}
	case "wf_act_freeze":
		if !positiveAt(stored.IntParams, 0) || stored.IntParams[0] > 86400 {
			return fmt.Errorf("%w: freeze duration", ErrInvalid)
		}
	case "wf_act_move_rotate_user":
		if len(stored.IntParams) < 2 || stored.IntParams[0] < 0 || stored.IntParams[0] > 7 || stored.IntParams[1] < 0 || stored.IntParams[1] > 7 {
			return fmt.Errorf("%w: actor movement", ErrInvalid)
		}
	case "wf_trg_user_performs_action", "wf_slc_users_byaction":
		if len(stored.IntParams) == 0 || stored.IntParams[0] < 1 || stored.IntParams[0] > 11 {
			return fmt.Errorf("%w: actor action", ErrInvalid)
		}
	case "wf_act_send_signal", "wf_trg_recv_signal":
		if strings.TrimSpace(stored.StringParam) == "" || len(stored.StringParam) > 64 {
			return fmt.Errorf("%w: signal", ErrInvalid)
		}
	case "wf_var_furni", "wf_var_user", "wf_var_room", "wf_var_reference",
		"wf_act_give_var", "wf_act_remove_var", "wf_act_change_var_val",
		"wf_cnd_has_var", "wf_cnd_neg_has_var", "wf_cnd_var_val_match",
		"wf_cnd_var_age_match", "wf_slc_furni_with_var", "wf_slc_users_with_var":
		if strings.TrimSpace(stored.StringParam) == "" || len(stored.StringParam) > 64 {
			return fmt.Errorf("%w: variable name", ErrInvalid)
		}
		if strings.HasPrefix(key, "wf_act_") || strings.HasPrefix(key, "wf_cnd_") {
			if len(stored.IntParams) == 0 || stored.IntParams[0] < 1 || stored.IntParams[0] > 4 {
				return fmt.Errorf("%w: variable scope", ErrInvalid)
			}
		}
		if key == "wf_var_reference" && !positiveAt(stored.IntParams, 0) {
			return fmt.Errorf("%w: reference room", ErrInvalid)
		}
	case "wf_cnd_slc_quantity", "wf_cnd_clock_matches":
		if len(stored.IntParams) < 2 || stored.IntParams[0] < 0 || stored.IntParams[0] > 5 {
			return fmt.Errorf("%w: comparison", ErrInvalid)
		}
	case "wf_xtra_filter_furni_by_var", "wf_xtra_filter_users_by_var":
		parts := strings.SplitN(strings.TrimSpace(stored.StringParam), "\t", 2)
		name := parts[0]
		if name == "" || len(name) > 64 ||
			len(stored.IntParams) < 3 ||
			stored.IntParams[0] < 0 || stored.IntParams[0] > 5 ||
			stored.IntParams[1] < 0 || stored.IntParams[1] > 1 ||
			stored.IntParams[2] < 0 || stored.IntParams[2] > 10000 {
			return fmt.Errorf("%w: variable filter", ErrInvalid)
		}
		if stored.IntParams[1] == 1 &&
			(len(parts) != 2 || strings.TrimSpace(parts[1]) == "" ||
				len(strings.TrimSpace(parts[1])) > 64 ||
				len(stored.IntParams) < 4 ||
				stored.IntParams[3] < 0 || stored.IntParams[3] > 3) {
			return fmt.Errorf("%w: variable filter amount source", ErrInvalid)
		}
	}
	return nil
}

// positiveAt reports whether a setting index stores a positive value.
func positiveAt(values []int32, index int) bool {
	return index >= 0 && index < len(values) && values[index] > 0
}
