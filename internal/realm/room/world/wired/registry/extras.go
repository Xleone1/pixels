package registry

// canonicalExtras returns selectors, game pickup, and highscore behavior.
func canonicalExtras() []Descriptor {
	return []Descriptor{
		{Key: "wf_xtra_random", Family: FamilyExtra, ClientCode: -1},
		{Key: "wf_xtra_unseen", Family: FamilyExtra, ClientCode: -1},
		{Key: "wf_xtra_or_eval", Family: FamilyExtra, ClientCode: -1},
		{Key: "wf_blob", Family: FamilyExtra, ClientCode: -1, Actor: ActorPlayer},
		{Key: "wf_highscore", Family: FamilyHighscore, ClientCode: -1},
	}
}

// canonicalSelectors returns the twenty official WIRED target selectors.
func canonicalSelectors() []Descriptor {
	return []Descriptor{
		{Key: "wf_slc_furni_area", Family: FamilySelector, ClientCode: 28, Editor: true},
		{Key: "wf_slc_furni_neighborhood", Family: FamilySelector, ClientCode: 29, Editor: true},
		{Key: "wf_slc_furni_bytype", Family: FamilySelector, ClientCode: 30, Selection: SelectionRequired, Editor: true},
		{Key: "wf_slc_users_area", Family: FamilySelector, ClientCode: 31, Editor: true},
		{Key: "wf_slc_users_neighborhood", Family: FamilySelector, ClientCode: 32, Editor: true},
		{Key: "wf_slc_furni_altitude", Family: FamilySelector, ClientCode: 44, Editor: true},
		{Key: "wf_slc_furni_onfurni", Family: FamilySelector, ClientCode: 45, Selection: SelectionRequired, Editor: true},
		{Key: "wf_slc_furni_picks", Family: FamilySelector, ClientCode: 46, Selection: SelectionRequired, Editor: true},
		{Key: "wf_slc_furni_signal", Family: FamilySelector, ClientCode: 47, Editor: true},
		{Key: "wf_slc_users_signal", Family: FamilySelector, ClientCode: 48, Editor: true},
		{Key: "wf_slc_users_bytype", Family: FamilySelector, ClientCode: 49, Editor: true},
		{Key: "wf_slc_users_team", Family: FamilySelector, ClientCode: 50, Editor: true},
		{Key: "wf_slc_users_byaction", Family: FamilySelector, ClientCode: 51, Editor: true},
		{Key: "wf_slc_users_byname", Family: FamilySelector, ClientCode: 52, Editor: true},
		{Key: "wf_slc_users_onfurni", Family: FamilySelector, ClientCode: 53, Selection: SelectionRequired, Editor: true},
		{Key: "wf_slc_users_group", Family: FamilySelector, ClientCode: 54, Editor: true},
		{Key: "wf_slc_users_handitem", Family: FamilySelector, ClientCode: 55, Editor: true},
		{Key: "wf_slc_furni_with_var", Family: FamilySelector, ClientCode: 75, Editor: true},
		{Key: "wf_slc_users_with_var", Family: FamilySelector, ClientCode: 76, Editor: true},
		{Key: "wf_slc_remote", Family: FamilySelector, ClientCode: 96, Selection: SelectionRequired, Editor: true},
	}
}

// canonicalVariables returns the four WIRED variable definitions in scope.
func canonicalVariables() []Descriptor {
	return []Descriptor{
		{Key: "wf_var_user", Family: FamilyVariable, ClientCode: 70, Editor: true},
		{Key: "wf_var_furni", Family: FamilyVariable, ClientCode: 71, Editor: true},
		{Key: "wf_var_room", Family: FamilyVariable, ClientCode: 72, Editor: true},
		{Key: "wf_var_reference", Family: FamilyVariable, ClientCode: 81, Editor: true},
	}
}

// canonicalAdditionalTriggers returns official WIRED trigger additions.
func canonicalAdditionalTriggers() []Descriptor {
	return []Descriptor{
		{Key: "wf_trg_recv_signal", Family: FamilyTrigger, ClientCode: 15, Actor: ActorOptional, Editor: true},
		{Key: "wf_trg_leave_room", Family: FamilyTrigger, ClientCode: 16, Actor: ActorPlayer, Editor: true},
		{Key: "wf_trg_user_performs_action", Family: FamilyTrigger, ClientCode: 21, Actor: ActorUnit, Editor: true},
		{Key: "wf_trg_clock_counter", Family: FamilyTrigger, ClientCode: 22, Actor: ActorOptional, Editor: true},
		{Key: "wf_trg_var_changed", Family: FamilyTrigger, ClientCode: 23, Actor: ActorOptional, Editor: true},
		{Key: "wf_trg_user_clicks_furni", Family: FamilyTrigger, ClientCode: 18, Selection: SelectionOptional, Actor: ActorPlayer, Editor: true},
		{Key: "wf_trg_user_clicks_tile", Family: FamilyTrigger, ClientCode: 19, Selection: SelectionOptional, Actor: ActorPlayer, Editor: true},
		{Key: "wf_trg_user_clicks_user", Family: FamilyTrigger, ClientCode: 20, Actor: ActorPlayer, Editor: true},
	}
}

// canonicalAdditionalEffects returns official WIRED effect additions.
func canonicalAdditionalEffects() []Descriptor {
	return []Descriptor{
		{Key: "wf_act_send_signal", Family: FamilyEffect, ClientCode: 33, Actor: ActorOptional, Editor: true},
		{Key: "wf_act_freeze", Family: FamilyEffect, ClientCode: 34, Actor: ActorUnit, Editor: true},
		{Key: "wf_act_unfreeze", Family: FamilyEffect, ClientCode: 35, Actor: ActorUnit, Editor: true},
		{Key: "wf_act_furni_to_user", Family: FamilyEffect, ClientCode: 36, Selection: SelectionRequired, Actor: ActorUnit, Editor: true},
		{Key: "wf_act_user_to_furni", Family: FamilyEffect, ClientCode: 37, Selection: SelectionRequired, Actor: ActorUnit, Editor: true},
		{Key: "wf_act_furni_to_furni", Family: FamilyEffect, ClientCode: 38, Selection: SelectionRequired, Editor: true},
		{Key: "wf_act_set_altitude", Family: FamilyEffect, ClientCode: 39, Selection: SelectionRequired, Editor: true},
		{Key: "wf_act_reset_furni", Family: FamilyEffect, ClientCode: 40, Selection: SelectionRequired, Editor: true},
		{Key: "wf_act_control_clock", Family: FamilyEffect, ClientCode: 41, Editor: true},
		{Key: "wf_act_control_clock_counter", Family: FamilyEffect, ClientCode: 42, Editor: true, Aliases: []string{"wf_act_adjust_clock"}},
		{Key: "wf_act_move_rotate_user", Family: FamilyEffect, ClientCode: 43, Actor: ActorUnit, Editor: true},
		{Key: "wf_act_give_var", Family: FamilyEffect, ClientCode: 69, Selection: SelectionOptional, Actor: ActorOptional, Editor: true},
		{Key: "wf_act_remove_var", Family: FamilyEffect, ClientCode: 73, Selection: SelectionOptional, Actor: ActorOptional, Editor: true},
		{Key: "wf_act_change_var_val", Family: FamilyEffect, ClientCode: 74, Selection: SelectionOptional, Actor: ActorOptional, Editor: true},
	}
}

// canonicalAdditionalConditions returns official WIRED condition additions.
func canonicalAdditionalConditions() []Descriptor {
	return []Descriptor{
		{Key: "wf_cnd_clock_matches", Family: FamilyCondition, ClientCode: 27, Editor: true},
		{Key: "wf_cnd_user_performs_action", Family: FamilyCondition, ClientCode: 28, Actor: ActorUnit, Editor: true},
		{Key: "wf_cnd_has_altitude", Family: FamilyCondition, ClientCode: 29, Selection: SelectionRequired, Editor: true},
		{Key: "wf_cnd_not_has_altitude", Family: FamilyCondition, ClientCode: 29, Selection: SelectionRequired, Editor: true},
		{Key: "wf_cnd_not_user_performs_action", Family: FamilyCondition, ClientCode: 30, Actor: ActorUnit, Editor: true},
		{Key: "wf_cnd_not_has_handitem", Family: FamilyCondition, ClientCode: 31, Actor: ActorPlayer, Editor: true},
		{Key: "wf_cnd_team_has_rank", Family: FamilyCondition, ClientCode: 35, Actor: ActorOptional, Editor: true},
		{Key: "wf_cnd_actor_dir", Family: FamilyCondition, ClientCode: 38, Actor: ActorUnit, Editor: true},
		{Key: "wf_cnd_slc_quantity", Family: FamilyCondition, ClientCode: 39, Editor: true},
		{Key: "wf_cnd_has_var", Family: FamilyCondition, ClientCode: 40, Actor: ActorOptional, Editor: true},
		{Key: "wf_cnd_neg_has_var", Family: FamilyCondition, ClientCode: 41, Actor: ActorOptional, Editor: true},
		{Key: "wf_cnd_var_val_match", Family: FamilyCondition, ClientCode: 42, Actor: ActorOptional, Editor: true},
		{Key: "wf_cnd_var_age_match", Family: FamilyCondition, ClientCode: 43, Actor: ActorOptional, Editor: true},
	}
}

// canonicalAdditionalExtras returns official WIRED stack modifiers.
func canonicalAdditionalExtras() []Descriptor {
	return []Descriptor{
		{Key: "wf_xtra_exec_in_order", Family: FamilyExtra, ClientCode: 64, Editor: true},
		{Key: "wf_xtra_filter_furni_by_var", Family: FamilyExtra, ClientCode: 78, Selection: SelectionOptional, Editor: true},
		{Key: "wf_xtra_filter_users_by_var", Family: FamilyExtra, ClientCode: 77, Selection: SelectionOptional, Editor: true},
	}
}

// CanonicalManifest returns the audited classic and WIRED behavior descriptors.
func CanonicalManifest() []Descriptor {
	manifest := make([]Descriptor, 0, 138)
	manifest = append(manifest, canonicalTriggers()...)
	manifest = append(manifest, canonicalEffects()...)
	manifest = append(manifest, canonicalConditions()...)
	manifest = append(manifest, canonicalExtras()...)
	manifest = append(manifest, canonicalSelectors()...)
	manifest = append(manifest, canonicalVariables()...)
	manifest = append(manifest, canonicalAdditionalTriggers()...)
	manifest = append(manifest, canonicalAdditionalEffects()...)
	manifest = append(manifest, canonicalAdditionalConditions()...)
	manifest = append(manifest, canonicalAdditionalExtras()...)
	return manifest
}

// CompatibilityManifest returns explicitly designed non-master extensions.
func CompatibilityManifest() []Descriptor {
	return []Descriptor{
		{Key: "wf_xtra_projectile", Family: FamilyExtra, ClientCode: 99, Selection: SelectionRequired, Actor: ActorOptional, Editor: true},
		{Key: "wf_cnd_valid_moves", Family: FamilyCondition, ClientCode: -1, Actor: ActorOptional},
		{Key: "wf_act_reset_highscore", Family: FamilyEffect, ClientCode: 0, Selection: SelectionRequired, Editor: true, Aliases: []string{"wf_cstm_reset_highscore"}},
		{Key: "wf_act_progress_achievement", Family: FamilyEffect, ClientCode: 7, Actor: ActorPlayer, Editor: true, Aliases: []string{"wf_cstm_achievement"}},
		{Key: "wf_act_progress_quest", Family: FamilyEffect, ClientCode: 7, Actor: ActorPlayer, Editor: true, Aliases: []string{"wf_cstm_progress_quest"}},
		{Key: "wf_act_start_quest", Family: FamilyEffect, ClientCode: 7, Actor: ActorPlayer, Editor: true, Aliases: []string{"wf_cstm_start_quest"}},
	}
}
