--;===========================================================
--; DEFAULT VALUES
--;===========================================================
local verInfo = main.f_fileRead("external/script/version", "r")

--This pre-made table (3/4 of the whole file) contains all default values used in screenpack. New table from parsed DEF file is merged on top of this one.
--This is important because there are more params available in Ikemen. Whole screenpack code refers to these values.
local motif =
{
	def = main.motifDef,
	info =
	{
		name = 'Default',
		author = 'Elecbyte',
		versiondate = {09, 01, 2009},
		mugenversion = '1.0',
		localcoord = {320, 240},
	},
	files =
	{
		spr = 'data/system.sff',
		snd = 'data/system.snd',
		logo_storyboard = '',
		intro_storyboard = '',
		select = 'data/select.def',
		fight = 'data/fight.def',
		font =
		{
			[1] = 'f-4x6.fnt',
			[2] = 'f-6x9.def',
			[3] = 'jg.fnt',
		},
		font_height = {},
		glyphs = 'data/glyphs.sff', --Ikemen feature
		module = '', --Ikemen feature
	},
	ja_files = {}, --not used in Ikemen
	music =
	{
		title_bgm = '',
		title_bgm_volume = 100,
		title_bgm_loop = 1,
		title_bgm_loopstart = 0,
		title_bgm_loopend = 0,
		select_bgm = '',
		select_bgm_volume = 100,
		select_bgm_loop = 1,
		select_bgm_loopstart = 0,
		select_bgm_loopend = 0,
		vs_bgm = '',
		vs_bgm_volume = 100,
		vs_bgm_loop = 1,
		vs_bgm_loopstart = 0,
		vs_bgm_loopend = 0,
		victory_bgm = '',
		victory_bgm_volume = 100,
		victory_bgm_loop = 1,
		victory_bgm_loopstart = 0,
		victory_bgm_loopend = 0,
		option_bgm = '', --Ikemen feature
		option_bgm_volume = 100, --Ikemen feature
		option_bgm_loop = 1, --Ikemen feature
		option_bgm_loopstart = 0, --Ikemen feature
		option_bgm_loopend = 0, --Ikemen feature
		replay_bgm = '', --Ikemen feature
		replay_bgm_volume = 100, --Ikemen feature
		replay_bgm_loop = 1, --Ikemen feature
		replay_bgm_loopstart = 0, --Ikemen feature
		replay_bgm_loopend = 0, --Ikemen feature
		continue_bgm = '', --Ikemen feature
		continue_bgm_volume = 100, --Ikemen feature
		continue_bgm_loop = 1, --Ikemen feature
		continue_bgm_loopstart = 0, --Ikemen feature
		continue_bgm_loopend = 0, --Ikemen feature
		continue_end_bgm = '', --Ikemen feature
		continue_end_bgm_volume = 100, --Ikemen feature
		continue_end_bgm_loop = 0, --Ikemen feature
		continue_end_bgm_loopstart = 0, --Ikemen feature
		continue_end_bgm_loopend = 0, --Ikemen feature
		results_bgm = '', --Ikemen feature
		results_bgm_volume = 100, --Ikemen feature
		results_bgm_loop = 1, --Ikemen feature
		results_bgm_loopstart = 0, --Ikemen feature
		results_bgm_loopend = 0, --Ikemen feature
		results_lose_bgm = '', --Ikemen feature
		results_lose_bgm_volume = 100, --Ikemen feature
		results_lose_bgm_loop = 1, --Ikemen feature
		results_lose_bgm_loopstart = 0, --Ikemen feature
		results_lose_bgm_loopend = 0, --Ikemen feature
		hiscore_bgm = '', --Ikemen feature
		hiscore_bgm_volume = 100, --Ikemen feature
		hiscore_bgm_loop = 1, --Ikemen feature
		hiscore_bgm_loopstart = 0, --Ikemen feature
		hiscore_bgm_loopend = 0, --Ikemen feature
	},
	title_info =
	{
		fadein_time = 10,
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadein_anim = -1, --Ikemen feature
		fadeout_time = 10,
		fadeout_col = {0, 0, 0}, --Ikemen feature
		fadeout_anim = -1, --Ikemen feature
		title_offset = {159, 15}, --Ikemen feature
		title_font = {-1, 0, 0, 255, 255, 255, -1}, --Ikemen feature
		title_scale = {1.0, 1.0}, --Ikemen feature
		title_text = 'MAIN MENU', --Ikemen feature
		loading_offset = {main.SP_Localcoord[1] - 1 - main.f_round(10 * main.SP_Localcoord[1] / 320), main.SP_Localcoord[2] - 8}, --Ikemen feature
		loading_font = {'default-3x5.def', 0, -1, 191, 191, 191, -1}, --Ikemen feature
		loading_scale = {1.0, 1.0}, --Ikemen feature
		loading_text = 'LOADING...', --Ikemen feature
		footer1_offset = {main.f_round(2 * main.SP_Localcoord[1] / 320), main.SP_Localcoord[2]}, --Ikemen feature
		footer1_font = {'default-3x5.def', 0, 1, 191, 191, 191, -1}, --Ikemen feature
		footer1_scale = {1.0, 1.0}, --Ikemen feature
		footer1_text = 'I.K.E.M.E.N. GO', --Ikemen feature
		footer2_offset = {main.SP_Localcoord[1] / 2, main.SP_Localcoord[2]}, --Ikemen feature
		footer2_font = {'default-3x5.def', 0, 0, 191, 191, 191, -1}, --Ikemen feature
		footer2_scale = {1.0, 1.0}, --Ikemen feature
		footer2_text = 'Press F1 for info', --Ikemen feature
		footer3_offset = {main.SP_Localcoord[1] - 1 - main.f_round(2 * main.SP_Localcoord[1] / 320), main.SP_Localcoord[2]}, --Ikemen feature
		footer3_font = {'default-3x5.def', 0, -1, 191, 191, 191, -1}, --Ikemen feature
		footer3_scale = {1.0, 1.0}, --Ikemen feature
		footer3_text = verInfo, --Ikemen feature
		footer_overlay_window = {0, main.SP_Localcoord[2] - 7, main.SP_Localcoord[1] - 1, main.SP_Localcoord[2] - 1}, --Ikemen feature
		footer_overlay_col = {0, 0, 64}, --Ikemen feature
		footer_overlay_alpha = {255, 100}, --Ikemen feature
		connecting_offset = {main.f_round(10 * main.SP_Localcoord[1] / 320), 40}, --Ikemen feature
		connecting_font = {'f-6x9.def', 0, 1, 255, 255, 255, -1}, --Ikemen feature
		connecting_scale = {1.0, 1.0}, --Ikemen feature
		connecting_host_text = 'Waiting for player 2... (%s)', --Ikemen feature
		connecting_join_text = 'Now connecting to %s... (%s)', --Ikemen feature
		connecting_overlay_window = {0, 0, main.SP_Localcoord[1], main.SP_Localcoord[2]}, --Ikemen feature
		connecting_overlay_col = {0, 0, 0}, --Ikemen feature
		connecting_overlay_alpha = {0, 128}, --Ikemen feature
		textinput_offset = {25, 32}, --Ikemen feature
		textinput_font = {'default-3x5.def', 0, 1, 191, 191, 191, -1}, --Ikemen feature
		textinput_scale = {1.0, 1.0}, --Ikemen feature
		textinput_name_text = 'Enter Host display name, e.g. John.\nExisting entries can be removed with DELETE button.', --Ikemen feature
		textinput_address_text = 'Enter Host IP address, e.g. 127.0.0.1\nCopied text can be pasted with INSERT button.', --Ikemen feature
		textinput_overlay_window = {0, 0, main.SP_Localcoord[1], main.SP_Localcoord[2]}, --Ikemen feature
		textinput_overlay_col = {0, 0, 0}, --Ikemen feature
		textinput_overlay_alpha = {0, 128}, --Ikemen feature
		menu_next_key = '$D&$F', --Ikemen feature
		menu_previous_key = '$U&$B', --Ikemen feature
		menu_accept_key = 'a&b&c&x&y&z&s', --Ikemen feature
		menu_hiscore_key = 's', --Ikemen feature
		menu_pos = {159, 158},
		--menu_bg_<itemname>_anim = -1, --Ikemen feature
		--menu_bg_<itemname>_spr = {}, --Ikemen feature
		--menu_bg_<itemname>_offset = {0, 0}, --Ikemen feature
		--menu_bg_<itemname>_facing = 1, --Ikemen feature
		--menu_bg_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		--menu_bg_active_<itemname>_anim = -1, --Ikemen feature
		--menu_bg_active_<itemname>_spr = {}, --Ikemen feature
		--menu_bg_active_<itemname>_offset = {0, 0}, --Ikemen feature
		--menu_bg_active_<itemname>_facing = 1, --Ikemen feature
		--menu_bg_active_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_offset = {0, 0}, --Ikemen feature
		menu_item_font = {-1, 0, 0, 191, 191, 191, -1},
		menu_item_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_active_offset = {0, 0}, --Ikemen feature
		menu_item_active_font = {-1, 0, 0, 255, 255, 255, -1},
		menu_item_active_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_spacing = {0, 13},
		menu_window_margins_y = {12, 8},
		menu_window_visibleitems = 5,
		menu_boxcursor_visible = 1,
		menu_boxcursor_coords = {-40, -10, 39, 2},
		menu_boxcursor_col = {255, 255, 255}, --Ikemen feature
		menu_boxcursor_alpharange = {10, 40, 2, 255, 255, 0}, --Ikemen feature
		menu_boxbg_visible = 0, --Ikemen feature
		menu_boxbg_col = {0, 0, 0}, --Ikemen feature
		menu_boxbg_alpha = {0, 128}, --Ikemen feature
		menu_arrow_up_anim = -1, --Ikemen feature
		menu_arrow_up_spr = {}, --Ikemen feature
		menu_arrow_up_offset = {0, 0}, --Ikemen feature
		menu_arrow_up_facing = 1, --Ikemen feature
		menu_arrow_up_scale = {1.0, 1.0}, --Ikemen feature
		menu_arrow_down_anim = -1, --Ikemen feature
		menu_arrow_down_spr = {}, --Ikemen feature
		menu_arrow_down_offset = {0, 0}, --Ikemen feature
		menu_arrow_down_facing = 1, --Ikemen feature
		menu_arrow_down_scale = {1.0, 1.0}, --Ikemen feature
		menu_title_uppercase = 1, --Ikemen feature
		cursor_move_snd = {100, 0},
		cursor_done_snd = {100, 1},
		cancel_snd = {100, 2},
		--cursor_<itemname>_snd = {-1, 0}, --Ikemen feature
		--menu_unlock_<itemname> = 'true', --Ikemen feature
		--menu_itemname_arcade = 'ARCADE',
		--menu_itemname_teamarcade = 'TEAM ARCADE',
		--menu_itemname_teamcoop = 'TEAM CO-OP',
		--menu_itemname_versus = 'VS MODE',
		--menu_itemname_teamversus = 'TEAM VERSUS',
		--menu_itemname_versuscoop = 'VERSUS CO-OP', --Ikemen feature
		--menu_itemname_freebattle = 'QUICK MATCH', --Ikemen feature
		--menu_itemname_storymode = 'STORY MODE', --Ikemen feature
		--menu_itemname_serverhost = 'HOST GAME', --Ikemen feature
		--menu_itemname_serverjoin = 'JOIN GAME', --Ikemen feature
		--menu_itemname_joinadd = 'NEW ADDRESS', --Ikemen feature
		--menu_itemname_netplayversus = 'VERSUS 2P', --Ikemen feature
		--menu_itemname_netplayteamcoop = 'ARCADE CO-OP', --Ikemen feature
		--menu_itemname_netplaysurvivalcoop = 'SURVIVAL CO-OP', --Ikemen feature
		--menu_itemname_training = 'TRAINING',
		--menu_itemname_trials = 'TRIALS', --Ikemen feature (not implemented yet)
		--menu_itemname_timeattack = 'TIME ATTACK', --Ikemen feature
		--menu_itemname_survival = 'SURVIVAL',
		--menu_itemname_survivalcoop = 'SURVIVAL CO-OP',
		--menu_itemname_bonusgames = 'BONUS GAMES', --Ikemen feature
		--menu_itemname_watch = 'CPU MATCH',
		--menu_itemname_randomtest = 'RANDOMTEST', --Ikemen feature
		--menu_itemname_replay = 'REPLAY', --Ikemen feature
		--menu_itemname_options = 'OPTIONS',
		--menu_itemname_back = 'BACK', --Ikemen feature
		--menu_itemname_exit = 'EXIT',
	},
	titlebgdef =
	{
		spr = '',
		bgclearcolor = {0, 0, 0},
	},
	infobox =
	{
		title_offset = {159, 15}, --Ikemen feature
		title_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		title_scale = {1.0, 1.0}, --Ikemen feature
		title_text = '', --Ikemen feature
		text_offset = {25, 32}, --Ikemen feature
		text_font = {'default-3x5.def', 0, 1, 191, 191, 191, -1},
		text_scale = {1.0, 1.0}, --Ikemen feature
		overlay_window = {0, 0, main.SP_Localcoord[1], main.SP_Localcoord[2]}, --Ikemen feature
		overlay_col = {0, 0, 0}, --Ikemen feature
		overlay_alpha = {0, 128}, --Ikemen feature
	},
	infobox_text = "Welcome to SUEHIRO's I.K.E.M.E.N GO engine!\n\n* This is a public development release, for testing purposes.\n* This build may contain bugs and incomplete features.\n* Your help and cooperation are appreciated!\n* Ikemen GO engine repositories: https://github.com/ikemen-engine\n* Original repo source code: https://osdn.net/users/supersuehiro/",
	ja_infobox_text = "", --not used in Ikemen
	select_info =
	{
		fadein_time = 10,
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadein_anim = -1, --Ikemen feature
		fadeout_time = 10,
		fadeout_col = {0, 0, 0}, --Ikemen feature
		fadeout_anim = -1, --Ikemen feature
		rows = 2,
		columns = 5,
		wrapping = 0,
		pos = {90, 170},
		showemptyboxes = 0,
		moveoveremptyboxes = 0,
		searchemptyboxesup = 0, --Ikemen feature
		searchemptyboxesdown = 0, --Ikemen feature
		cell_size = {27, 27},
		cell_spacing = {2, 2}, --Mugen/Ikemen feature (in Mugen spacing x value is used for both coordinates)
		cell_bg_anim = -1,
		cell_bg_spr = {},
		cell_bg_offset = {0, 0},
		cell_bg_facing = 1,
		cell_bg_scale = {1.0, 1.0},
		cell_random_anim = -1,
		cell_random_spr = {},
		cell_random_offset = {0, 0},
		cell_random_facing = 1,
		cell_random_scale = {1.0, 1.0},
		cell_random_switchtime = 4,
		--cell_<col>_<row>_offset = {0, 0}, --Ikemen feature
		--cell_<col>_<row>_facing = 1, --Ikemen feature
		--cell_<col>_<row>_skip = 0, --Ikemen feature
		p1_cursor_startcell = {0, 0},
		p1_cursor_active_anim = -1,
		p1_cursor_active_spr = {},
		p1_cursor_active_offset = {0, 0},
		p1_cursor_active_facing = 1,
		p1_cursor_active_scale = {1.0, 1.0},
		p1_cursor_done_anim = -1,
		p1_cursor_done_spr = {},
		p1_cursor_done_offset = {0, 0},
		p1_cursor_done_facing = 1,
		p1_cursor_done_scale = {1.0, 1.0},
		p1_cursor_move_snd = {100, 0},
		p1_cursor_done_snd = {100, 1},
		p1_random_move_snd = {100, 0},
		p2_cursor_startcell = {0, 4},
		p2_cursor_active_anim = -1,
		p2_cursor_active_spr = {},
		p2_cursor_active_offset = {0, 0},
		p2_cursor_active_facing = 1,
		p2_cursor_active_scale = {1.0, 1.0},
		p2_cursor_done_anim = -1,
		p2_cursor_done_spr = {},
		p2_cursor_done_offset = {0, 0},
		p2_cursor_done_facing = 1,
		p2_cursor_done_scale = {1.0, 1.0},
		p2_cursor_blink = 1,
		p2_cursor_switchtime = 3, --Ikemen feature
		p2_cursor_move_snd = {100, 0},
		p2_cursor_done_snd = {100, 1},
		p2_random_move_snd = {100, 0},
		--p<pn>_cursor_startcell = {0, 0}, --Ikemen feature
		--p<pn>_cursor_active_anim = -1, --Ikemen feature
		--p<pn>_cursor_active_spr = {}, --Ikemen feature
		--p<pn>_cursor_active_offset = {0, 0}, --Ikemen feature
		--p<pn>_cursor_active_facing = 1, --Ikemen feature
		--p<pn>_cursor_active_scale = {1.0, 1.0}, --Ikemen feature
		--p<pn>_cursor_done_anim = -1, --Ikemen feature
		--p<pn>_cursor_done_spr = {}, --Ikemen feature
		--p<pn>_cursor_done_offset = {0, 0}, --Ikemen feature
		--p<pn>_cursor_done_facing = 1, --Ikemen feature
		--p<pn>_cursor_done_scale = {1.0, 1.0}, --Ikemen feature
		--p<pn>_cursor_active_<col>_<row>_anim = -1, --Ikemen feature
		--p<pn>_cursor_active_<col>_<row>_spr = {}, --Ikemen feature
		--p<pn>_cursor_active_<col>_<row>_offset = {0, 0}, --Ikemen feature
		--p<pn>_cursor_active_<col>_<row>_facing = 1, --Ikemen feature
		--p<pn>_cursor_active_<col>_<row>_scale = {1.0, 1.0}, --Ikemen feature
		--p<pn>_cursor_done_<col>_<row>_anim = -1, --Ikemen feature
		--p<pn>_cursor_done_<col>_<row>_spr = {}, --Ikemen feature
		--p<pn>_cursor_done_<col>_<row>_offset = {0, 0}, --Ikemen feature
		--p<pn>_cursor_done_<col>_<row>_facing = 1, --Ikemen feature
		--p<pn>_cursor_done_<col>_<row>_scale = {1.0, 1.0}, --Ikemen feature
		--p<pn>_cursor_move_snd = {100, 0}, --Ikemen feature
		--p<pn>_cursor_done_snd = {100, 1}, --Ikemen feature
		--p<pn>_random_move_snd = {100, 0}, --Ikemen feature
		random_move_snd_cancel = 0,
		stage_move_snd = {100, 0},
		stage_done_snd = {100, 1},
		cancel_snd = {100, 2},
		portrait_anim = -1, --Ikemen feature
		portrait_spr = {9000, 0},
		portrait_offset = {0, 0},
		portrait_facing = 1,
		portrait_scale = {1.0, 1.0},
		title_offset = {0, 0},
		title_font = {-1, 0, 0, 255, 255, 255, -1},
		title_scale = {1.0, 1.0}, --Ikemen feature
		title_arcade_text = 'Arcade', --Ikemen feature
		title_teamarcade_text = 'Team Arcade', --Ikemen feature
		title_teamcoop_text = 'Team Cooperative', --Ikemen feature
		title_versus_text = 'Versus Mode', --Ikemen feature
		title_teamversus_text = 'Team Versus', --Ikemen feature
		title_versuscoop_text = 'Versus Cooperative', --Ikemen feature
		title_freebattle_text = 'Quick Match', --Ikemen feature
		title_storymode_text = 'Story Mode', --Ikemen feature
		title_netplayversus_text = 'Online Versus', --Ikemen feature
		title_netplayteamcoop_text = 'Online Cooperative', --Ikemen feature
		title_netplaysurvivalcoop_text = 'Online Survival', --Ikemen feature
		title_training_text = 'Training Mode', --Ikemen feature
		title_timeattack_text = 'Time Attack', --Ikemen feature
		title_survival_text = 'Survival', --Ikemen feature
		title_survivalcoop_text = 'Survival Cooperative', --Ikemen feature
		title_bonus_text = 'Bonus', --Ikemen feature
		title_watch_text = 'Watch Mode', --Ikemen feature
		--title_replay_text = 'Replay', --Ikemen feature
		p1_face_pos = {0, 0},
		p1_face_num = 1, --Ikemen feature
		p1_face_anim = -1, --Ikemen feature
		p1_face_spr = {9000, 1},
		p1_face_done_anim = -1, --Ikemen feature
		p1_face_done_spr = {9000, 1}, --Ikemen feature
		p1_face_offset = {0, 0},
		p1_face_facing = 1,
		p1_face_scale = {1.0, 1.0},
		p1_face_window = {},
		p1_face_spacing = {0, 0}, --Ikemen feature
		p1_face_padding = 0, --Ikemen feature
		p2_face_pos = {0, 0},
		p2_face_num = 1, --Ikemen feature
		p2_face_anim = -1, --Ikemen feature
		p2_face_done_anim = -1, --Ikemen feature
		p2_face_done_spr = {9000, 1}, --Ikemen feature
		p2_face_spr = {9000, 1},
		p2_face_offset = {0, 0},
		p2_face_facing = -1,
		p2_face_scale = {1.0, 1.0},
		p2_face_window = {},
		p2_face_spacing = {0, 0}, --Ikemen feature
		p2_face_padding = 0, --Ikemen feature
		--p<pn>_member<num>_face_anim = -1, --Ikemen feature
		--p<pn>_member<num>_face_spr = {9000, 1}, --Ikemen feature
		--p<pn>_member<num>_face_done_anim = -1, --Ikemen feature
		--p<pn>_member<num>_face_done_spr = {9000, 1}, --Ikemen feature
		--p<pn>_member<num>_face_offset = {0, 0}, --Ikemen feature
		--p<pn>_member<num>_face_scale = {1.0, 1.0}, --Ikemen feature
		--p<pn>_member<num>_face_slide_speed = {0, 0}, --Ikemen feature
		--p<pn>_member<num>_face_slide_dist = {0, 0}, --Ikemen feature
		p1_face2_anim = -1, --Ikemen feature
		p1_face2_spr = {}, --Ikemen feature
		p1_face2_offset = {0, 0}, --Ikemen feature
		p1_face2_facing = 1, --Ikemen feature
		p1_face2_scale = {1.0, 1.0}, --Ikemen feature
		p1_face2_window = {}, --Ikemen feature
		p2_face2_anim = -1, --Ikemen feature
		p2_face2_spr = {}, --Ikemen feature
		p2_face2_offset = {0, 0}, --Ikemen feature
		p2_face2_facing = -1, --Ikemen feature
		p2_face2_scale = {1.0, 1.0}, --Ikemen feature
		p2_face2_window = {}, --Ikemen feature
		p1_name_num = 4, --Ikemen feature
		p1_name_offset = {0, 0},
		p1_name_font = {-1, 4, 1, 255, 255, 255, -1},
		p1_name_scale = {1.0, 1.0}, --Ikemen feature
		p1_name_spacing = {0, 14},
		p1_name_random_text = 'Random', --Ikemen feature
		p2_name_num = 4, --Ikemen feature
		p2_name_offset = {0, 0},
		p2_name_font = {-1, 1, -1, 255, 255, 255, -1},
		p2_name_scale = {1.0, 1.0}, --Ikemen feature
		p2_name_spacing = {0, 14},
		p2_name_random_text = 'Random', --Ikemen feature
		stage_pos = {0, 0},
		stage_active_offset = {0, 0}, --Ikemen feature
		stage_active_font = {-1, 0, 0, 255, 255, 255, -1},
		stage_active_scale = {1.0, 1.0}, --Ikemen feature
		stage_active_switchtime = 2, --Ikemen feature
		stage_active2_offset = {0, 0}, --Ikemen feature
		stage_active2_font = {-1, 0, 0, 255, 255, 255, -1},
		stage_active2_scale = {1.0, 1.0}, --Ikemen feature
		stage_done_offset = {0, 0}, --Ikemen feature
		stage_done_font = {-1, 0, 0, 255, 255, 255, -1},
		stage_done_scale = {1.0, 1.0}, --Ikemen feature
		stage_text = 'Stage %i: %s', --Ikemen feature
		stage_random_text = 'Stage: Random', --Ikemen feature
		stage_portrait_anim = -1, --Ikemen feature
		stage_portrait_spr = {}, --Ikemen feature
		stage_portrait_offset = {0, 0}, --Ikemen feature
		stage_portrait_scale = {1.0, 1.0}, --Ikemen feature
		stage_portrait_bg_anim = -1, --Ikemen feature
		stage_portrait_bg_spr = {}, --Ikemen feature
		stage_portrait_bg_offset = {0, 0}, --Ikemen feature
		stage_portrait_bg_scale = {1.0, 1.0}, --Ikemen feature
		stage_portrait_random_anim = -1, --Ikemen feature
		stage_portrait_random_spr = {}, --Ikemen feature
		stage_portrait_random_offset = {0, 0}, --Ikemen feature
		stage_portrait_random_scale = {1.0, 1.0}, --Ikemen feature
		stage_portrait_window = {}, --Ikemen feature
		teammenu_move_wrapping = 1,
		teammenu_itemname_single = 'Single', --Ikemen feature
		teammenu_itemname_simul = 'Simul', --Ikemen feature
		teammenu_itemname_turns = 'Turns', --Ikemen feature
		teammenu_itemname_tag = '', --Ikemen feature (Tag)
		teammenu_itemname_ratio = '', --Ikemen feature (Ratio)
		--teammenu_itemname_<gamemode>_<teammode> = '', --Ikemen feature
		p1_teammenu_pos = {0, 0},
		p1_teammenu_bg_anim = -1,
		p1_teammenu_bg_spr = {},
		p1_teammenu_bg_offset = {0, 0},
		p1_teammenu_bg_facing = 1,
		p1_teammenu_bg_scale = {1.0, 1.0},
		--p1_teammenu_bg_<itemname>_anim = -1, --Ikemen feature
		--p1_teammenu_bg_<itemname>_spr = {}, --Ikemen feature
		--p1_teammenu_bg_<itemname>_offset = {0, 0}, --Ikemen feature
		--p1_teammenu_bg_<itemname>_facing = 1, --Ikemen feature
		--p1_teammenu_bg_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		--p1_teammenu_bg_active_<itemname>_anim = -1, --Ikemen feature
		--p1_teammenu_bg_active_<itemname>_spr = {}, --Ikemen feature
		--p1_teammenu_bg_active_<itemname>_offset = {0, 0}, --Ikemen feature
		--p1_teammenu_bg_active_<itemname>_facing = 1, --Ikemen feature
		--p1_teammenu_bg_active_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		p1_teammenu_selftitle_anim = -1,
		p1_teammenu_selftitle_spr = {},
		p1_teammenu_selftitle_offset = {0, 0},
		p1_teammenu_selftitle_facing = 1,
		p1_teammenu_selftitle_scale = {1.0, 1.0},
		p1_teammenu_selftitle_font = {-1, 0, 1, 255, 255, 255, -1},
		p1_teammenu_selftitle_scale = {1.0, 1.0},
		p1_teammenu_selftitle_text = '',
		p1_teammenu_enemytitle_anim = -1,
		p1_teammenu_enemytitle_spr = {},
		p1_teammenu_enemytitle_offset = {0, 0},
		p1_teammenu_enemytitle_facing = 1,
		p1_teammenu_enemytitle_scale = {1.0, 1.0},
		p1_teammenu_enemytitle_font = {-1, 0, 1, 255, 255, 255, -1},
		p1_teammenu_enemytitle_scale = {1.0, 1.0},
		p1_teammenu_enemytitle_text = '',
		p1_teammenu_move_snd = {100, 0},
		p1_teammenu_value_snd = {100, 0},
		p1_teammenu_done_snd = {100, 1},
		p1_teammenu_item_offset = {0, 0},
		p1_teammenu_item_spacing = {0, 0},
		p1_teammenu_item_font = {-1, 0, 1, 255, 255, 255, -1},
		p1_teammenu_item_scale = {1.0, 1.0}, --Ikemen feature
		p1_teammenu_item_active_font = {-1, 3, 1, 255, 255, 255, -1},
		p1_teammenu_item_active_scale = {1.0, 1.0}, --Ikemen feature
		p1_teammenu_item_active_switchtime = 2, --Ikemen feature
		p1_teammenu_item_active2_font = {-1, 0, 1, 255, 255, 255, -1},
		p1_teammenu_item_active2_scale = {1.0, 1.0}, --Ikemen feature
		p1_teammenu_item_cursor_anim = -1,
		p1_teammenu_item_cursor_spr = {},
		p1_teammenu_item_cursor_offset = {0, 0},
		p1_teammenu_item_cursor_facing = 1,
		p1_teammenu_item_cursor_scale = {1.0, 1.0},
		p1_teammenu_value_icon_anim = -1,
		p1_teammenu_value_icon_spr = {},
		p1_teammenu_value_icon_offset = {0, 0},
		p1_teammenu_value_icon_facing = 1,
		p1_teammenu_value_icon_scale = {1.0, 1.0},
		p1_teammenu_value_empty_icon_anim = -1,
		p1_teammenu_value_empty_icon_spr = {},
		p1_teammenu_value_empty_icon_offset = {0, 0},
		p1_teammenu_value_empty_icon_facing = 1,
		p1_teammenu_value_empty_icon_scale = {1.0, 1.0},
		p1_teammenu_value_spacing = {6, 0},
		p1_teammenu_ratio1_icon_anim = -1, --Ikemen feature
		p1_teammenu_ratio1_icon_spr = {}, --Ikemen feature
		p1_teammenu_ratio1_icon_offset = {0, 0}, --Ikemen feature
		p1_teammenu_ratio1_icon_facing = 1, --Ikemen feature
		p1_teammenu_ratio1_icon_scale = {1.0, 1.0}, --Ikemen feature
		p1_teammenu_ratio2_icon_anim = -1, --Ikemen feature
		p1_teammenu_ratio2_icon_spr = {}, --Ikemen feature
		p1_teammenu_ratio2_icon_offset = {0, 0}, --Ikemen feature
		p1_teammenu_ratio2_icon_facing = 1, --Ikemen feature
		p1_teammenu_ratio2_icon_scale = {1.0, 1.0}, --Ikemen feature
		p1_teammenu_ratio3_icon_anim = -1, --Ikemen feature
		p1_teammenu_ratio3_icon_spr = {}, --Ikemen feature
		p1_teammenu_ratio3_icon_offset = {0, 0}, --Ikemen feature
		p1_teammenu_ratio3_icon_facing = 1, --Ikemen feature
		p1_teammenu_ratio3_icon_scale = {1.0, 1.0}, --Ikemen feature
		p1_teammenu_ratio4_icon_anim = -1, --Ikemen feature
		p1_teammenu_ratio4_icon_spr = {}, --Ikemen feature
		p1_teammenu_ratio4_icon_offset = {0, 0}, --Ikemen feature
		p1_teammenu_ratio4_icon_facing = 1, --Ikemen feature
		p1_teammenu_ratio4_icon_scale = {1.0, 1.0}, --Ikemen feature
		p1_teammenu_ratio5_icon_anim = -1, --Ikemen feature
		p1_teammenu_ratio5_icon_spr = {}, --Ikemen feature
		p1_teammenu_ratio5_icon_offset = {0, 0}, --Ikemen feature
		p1_teammenu_ratio5_icon_facing = 1, --Ikemen feature
		p1_teammenu_ratio5_icon_scale = {1.0, 1.0}, --Ikemen feature
		p1_teammenu_ratio6_icon_anim = -1, --Ikemen feature
		p1_teammenu_ratio6_icon_spr = {}, --Ikemen feature
		p1_teammenu_ratio6_icon_offset = {0, 0}, --Ikemen feature
		p1_teammenu_ratio6_icon_facing = 1, --Ikemen feature
		p1_teammenu_ratio6_icon_scale = {1.0, 1.0}, --Ikemen feature
		p1_teammenu_ratio7_icon_anim = -1, --Ikemen feature
		p1_teammenu_ratio7_icon_spr = {}, --Ikemen feature
		p1_teammenu_ratio7_icon_offset = {0, 0}, --Ikemen feature
		p1_teammenu_ratio7_icon_facing = 1, --Ikemen feature
		p1_teammenu_ratio7_icon_scale = {1.0, 1.0}, --Ikemen feature
		p2_teammenu_pos = {0, 0},
		p2_teammenu_bg_anim = -1,
		p2_teammenu_bg_spr = {},
		p2_teammenu_bg_offset = {0, 0},
		p2_teammenu_bg_facing = 1,
		p2_teammenu_bg_scale = {1.0, 1.0},
		--p2_teammenu_bg_<itemname>_anim = -1, --Ikemen feature
		--p2_teammenu_bg_<itemname>_spr = {}, --Ikemen feature
		--p2_teammenu_bg_<itemname>_offset = {0, 0}, --Ikemen feature
		--p2_teammenu_bg_<itemname>_facing = 1, --Ikemen feature
		--p2_teammenu_bg_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		--p2_teammenu_bg_active_<itemname>_anim = -1, --Ikemen feature
		--p2_teammenu_bg_active_<itemname>_spr = {}, --Ikemen feature
		--p2_teammenu_bg_active_<itemname>_offset = {0, 0}, --Ikemen feature
		--p2_teammenu_bg_active_<itemname>_facing = 1, --Ikemen feature
		--p2_teammenu_bg_active_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		p2_teammenu_selftitle_anim = -1,
		p2_teammenu_selftitle_spr = {},
		p2_teammenu_selftitle_offset = {0, 0},
		p2_teammenu_selftitle_facing = 1,
		p2_teammenu_selftitle_scale = {1.0, 1.0},
		p2_teammenu_selftitle_font = {-1, 0, -1, 255, 255, 255, -1},
		p2_teammenu_selftitle_scale = {1.0, 1.0},
		p2_teammenu_selftitle_text = '',
		p2_teammenu_enemytitle_anim = -1,
		p2_teammenu_enemytitle_spr = {},
		p2_teammenu_enemytitle_offset = {0, 0},
		p2_teammenu_enemytitle_facing = 1,
		p2_teammenu_enemytitle_scale = {1.0, 1.0},
		p2_teammenu_enemytitle_font = {-1, 0, -1, 255, 255, 255, -1},
		p2_teammenu_enemytitle_scale = {1.0, 1.0},
		p2_teammenu_enemytitle_text = '',
		p2_teammenu_move_snd = {100, 0},
		p2_teammenu_value_snd = {100, 0},
		p2_teammenu_done_snd = {100, 1},
		p2_teammenu_item_offset = {0, 0},
		p2_teammenu_item_spacing = {0, 0},
		p2_teammenu_item_font = {-1, 0, -1, 255, 255, 255, -1},
		p2_teammenu_item_scale = {1.0, 1.0}, --Ikemen feature
		p2_teammenu_item_active_font = {-1, 1, -1, 255, 255, 255, -1},
		p2_teammenu_item_active_scale = {1.0, 1.0}, --Ikemen feature
		p2_teammenu_item_active_switchtime = 2, --Ikemen feature
		p2_teammenu_item_active2_font = {-1, 0, -1, 255, 255, 255, -1},
		p2_teammenu_item_active2_scale = {1.0, 1.0}, --Ikemen feature
		p2_teammenu_item_cursor_anim = -1,
		p2_teammenu_item_cursor_spr = {},
		p2_teammenu_item_cursor_offset = {0, 0},
		p2_teammenu_item_cursor_facing = 1,
		p2_teammenu_item_cursor_scale = {1.0, 1.0},
		p2_teammenu_value_icon_anim = -1,
		p2_teammenu_value_icon_spr = {},
		p2_teammenu_value_icon_offset = {0, 0},
		p2_teammenu_value_icon_facing = 1,
		p2_teammenu_value_icon_scale = {1.0, 1.0},
		p2_teammenu_value_empty_icon_anim = -1,
		p2_teammenu_value_empty_icon_spr = {},
		p2_teammenu_value_empty_icon_offset = {0, 0},
		p2_teammenu_value_empty_icon_facing = 1,
		p2_teammenu_value_empty_icon_scale = {1.0, 1.0},
		p2_teammenu_value_spacing = {-6, 0},
		p2_teammenu_ratio1_icon_anim = -1, --Ikemen feature
		p2_teammenu_ratio1_icon_spr = {}, --Ikemen feature
		p2_teammenu_ratio1_icon_offset = {0, 0}, --Ikemen feature
		p2_teammenu_ratio1_icon_facing = 1, --Ikemen feature
		p2_teammenu_ratio1_icon_scale = {1.0, 1.0}, --Ikemen feature
		p2_teammenu_ratio2_icon_anim = -1, --Ikemen feature
		p2_teammenu_ratio2_icon_spr = {}, --Ikemen feature
		p2_teammenu_ratio2_icon_offset = {0, 0}, --Ikemen feature
		p2_teammenu_ratio2_icon_facing = 1, --Ikemen feature
		p2_teammenu_ratio2_icon_scale = {1.0, 1.0}, --Ikemen feature
		p2_teammenu_ratio3_icon_anim = -1, --Ikemen feature
		p2_teammenu_ratio3_icon_spr = {}, --Ikemen feature
		p2_teammenu_ratio3_icon_offset = {0, 0}, --Ikemen feature
		p2_teammenu_ratio3_icon_facing = 1, --Ikemen feature
		p2_teammenu_ratio3_icon_scale = {1.0, 1.0}, --Ikemen feature
		p2_teammenu_ratio4_icon_anim = -1, --Ikemen feature
		p2_teammenu_ratio4_icon_spr = {}, --Ikemen feature
		p2_teammenu_ratio4_icon_offset = {0, 0}, --Ikemen feature
		p2_teammenu_ratio4_icon_facing = 1, --Ikemen feature
		p2_teammenu_ratio4_icon_scale = {1.0, 1.0}, --Ikemen feature
		p2_teammenu_ratio5_icon_anim = -1, --Ikemen feature
		p2_teammenu_ratio5_icon_spr = {}, --Ikemen feature
		p2_teammenu_ratio5_icon_offset = {0, 0}, --Ikemen feature
		p2_teammenu_ratio5_icon_facing = 1, --Ikemen feature
		p2_teammenu_ratio5_icon_scale = {1.0, 1.0}, --Ikemen feature
		p2_teammenu_ratio6_icon_anim = -1, --Ikemen feature
		p2_teammenu_ratio6_icon_spr = {}, --Ikemen feature
		p2_teammenu_ratio6_icon_offset = {0, 0}, --Ikemen feature
		p2_teammenu_ratio6_icon_facing = 1, --Ikemen feature
		p2_teammenu_ratio6_icon_scale = {1.0, 1.0}, --Ikemen feature
		p2_teammenu_ratio7_icon_anim = -1, --Ikemen feature
		p2_teammenu_ratio7_icon_spr = {}, --Ikemen feature
		p2_teammenu_ratio7_icon_offset = {0, 0}, --Ikemen feature
		p2_teammenu_ratio7_icon_facing = 1, --Ikemen feature
		p2_teammenu_ratio7_icon_scale = {1.0, 1.0}, --Ikemen feature
		p1_teammenu_next_key = '$D', --Ikemen feature
		p1_teammenu_previous_key = '$U', --Ikemen feature
		p1_teammenu_add_key = '$F', --Ikemen feature
		p1_teammenu_subtract_key = '$B', --Ikemen feature
		p1_teammenu_accept_key = 'a&b&c&x&y&z&s', --Ikemen feature
		p2_teammenu_next_key = '$D', --Ikemen feature
		p2_teammenu_previous_key = '$U', --Ikemen feature
		p2_teammenu_add_key = '$B', --Ikemen feature
		p2_teammenu_subtract_key = '$F', --Ikemen feature
		p2_teammenu_accept_key = 'a&b&c&x&y&z&s', --Ikemen feature
		timer_offset = {0, 0}, --Ikemen feature
		timer_font = {-1, 0, 0, 255, 255, 255, -1}, --Ikemen feature
		timer_scale = {1.0, 1.0}, --Ikemen feature
		timer_text = '%i', --Ikemen feature
		timer_count = -1, --Ikemen feature
		timer_framespercount = 60, --Ikemen feature
		timer_displaytime = 10, --Ikemen feature
		record_offset = {0, 0}, --Ikemen feature
		record_font = {-1, 0, 0, 255, 255, 255, -1}, --Ikemen feature
		record_scale = {1.0, 1.0}, --Ikemen feature
		--record_<gamemode>_text = '', --Ikemen feature
		p1_swap_snd = {-1, 0}, --Ikemen feature
		p2_swap_snd = {-1, 0}, --Ikemen feature
		p1_select_snd = {-1, 0}, --Ikemen feature (data read from character SND)
		p2_select_snd = {-1, 0}, --Ikemen feature (data read from character SND)
	},
	selectbgdef =
	{
		spr = '',
		bgclearcolor = {0, 0, 0},
	},
	vs_screen =
	{
		orderselect_enabled = 0, --Ikemen feature
		done_time = 60, --Ikemen feature
		fadein_time = 15,
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadein_anim = -1, --Ikemen feature
		fadeout_time = 15,
		fadeout_col = {0, 0, 0}, --Ikemen feature
		fadeout_anim = -1, --Ikemen feature
		time = 150,
		match_text = 'Match %i',
		match_offset = {159, 12},
		match_font = {-1, 0, 0, 255, 255, 255, -1},
		match_scale = {1.0, 1.0},
		p1_pos = {0, 0},
		p1_num = 1, --Ikemen feature
		p1_anim = -1, --Ikemen feature
		p1_spr = {9000, 1},
		p1_offset = {0, 0},
		p1_facing = 1,
		p1_scale = {1.0, 1.0},
		p1_window = {},
		p1_spacing = {0, 0}, --Ikemen feature
		p1_padding = 0, --Ikemen feature
		p2_pos = {0, 0},
		p2_num = 1, --Ikemen feature
		p2_anim = -1, --Ikemen feature
		p2_spr = {9000, 1},
		p2_offset = {0, 0},
		p2_facing = -1,
		p2_scale = {1.0, 1.0},
		p2_window = {},
		p2_spacing = {0, 0}, --Ikemen feature
		p2_padding = 0, --Ikemen feature
		--p<pn>_member<num>_anim = -1, --Ikemen feature
		--p<pn>_member<num>_spr = {9000, 1}, --Ikemen feature
		--p<pn>_member<num>_offset = {0, 0}, --Ikemen feature
		--p<pn>_member<num>_scale = {1.0, 1.0}, --Ikemen feature
		--p<pn>_member<num>_slide_speed = {0, 0}, --Ikemen feature
		--p<pn>_member<num>_slide_dist = {0, 0}, --Ikemen feature
		p1_face2_anim = -1, --Ikemen feature
		p1_face2_spr = {}, --Ikemen feature
		p1_face2_offset = {0, 0}, --Ikemen feature
		p1_face2_facing = 1, --Ikemen feature
		p1_face2_scale = {1.0, 1.0}, --Ikemen feature
		p1_face2_window = {}, --Ikemen feature
		p2_face2_anim = -1, --Ikemen feature
		p2_face2_spr = {}, --Ikemen feature
		p2_face2_offset = {0, 0}, --Ikemen feature
		p2_face2_facing = -1, --Ikemen feature
		p2_face2_scale = {1.0, 1.0}, --Ikemen feature
		p2_face2_window = {}, --Ikemen feature
		p1_name_num = 4, --Ikemen feature
		p1_name_pos = {0, 0},
		p1_name_offset = {0, 0},
		p1_name_font = {-1, 0, 0, 255, 255, 255, -1},
		p1_name_scale = {1.0, 1.0}, --Ikemen feature
		p1_name_spacing = {0, 14},
		p2_name_num = 4, --Ikemen feature
		p2_name_pos = {0, 0},
		p2_name_offset = {0, 0},
		p2_name_font = {-1, 0, 0, 255, 255, 255, -1},
		p2_name_scale = {1.0, 1.0}, --Ikemen feature
		p2_name_spacing = {0, 14},
		--p<pn>_member<num>_key = "", --Ikemen feature
		p1_accept_key = "a&b&c&x&y&z&s", --Ikemen feature
		p1_skip_key = "s", --Ikemen feature
		p2_accept_key = "a&b&c&x&y&z&s", --Ikemen feature
		p2_skip_key = "s", --Ikemen feature
		--p<pn>_member<num>_icon_anim = -1, --Ikemen feature
		--p<pn>_member<num>_icon_spr = {}, --Ikemen feature
		--p<pn>_member<num>_icon_done_anim = -1, --Ikemen feature
		--p<pn>_member<num>_icon_done_spr = {}, --Ikemen feature
		--p<pn>_member<num>_icon_offset = {0, 0}, --Ikemen feature
		--p<pn>_member<num>_icon_facing = 1, --Ikemen feature
		--p<pn>_member<num>_icon_scale = {1.0, 1.0}, --Ikemen feature
		p1_value_icon_anim = -1, --Ikemen feature
		p1_value_icon_spr = {}, --Ikemen feature
		--p1_value_icon_member<num>_anim = -1, --Ikemen feature
		--p1_value_icon_member<num>_spr = {}, --Ikemen feature
		p1_value_icon_offset = {0, 0}, --Ikemen feature
		p1_value_icon_facing = 1, --Ikemen feature
		p1_value_icon_scale = {1.0, 1.0}, --Ikemen feature
		p1_value_empty_icon_anim = -1, --Ikemen feature
		p1_value_empty_icon_spr = {}, --Ikemen feature
		--p1_value_empty_icon_member<num>_anim = -1, --Ikemen feature
		--p1_value_empty_icon_member<num>_spr = {}, --Ikemen feature
		p1_value_empty_icon_offset = {0, 0}, --Ikemen feature
		p1_value_empty_icon_facing = 1, --Ikemen feature
		p1_value_empty_icon_scale = {1.0, 1.0}, --Ikemen feature
		p1_value_icon_spacing = {0, 0}, --Ikemen feature
		p1_value_snd = {-1, 0}, --Ikemen feature
		p2_value_icon_anim = -1, --Ikemen feature
		p2_value_icon_spr = {}, --Ikemen feature
		--p2_value_icon_member<num>_anim = -1, --Ikemen feature
		--p2_value_icon_member<num>_spr = {}, --Ikemen feature
		p2_value_icon_offset = {0, 0}, --Ikemen feature
		p2_value_icon_facing = 1, --Ikemen feature
		p2_value_icon_scale = {1.0, 1.0}, --Ikemen feature
		p2_value_empty_icon_anim = -1, --Ikemen feature
		p2_value_empty_icon_spr = {}, --Ikemen feature
		--p2_value_empty_icon_member<num>_anim = -1, --Ikemen feature
		--p2_value_empty_icon_member<num>_spr = {}, --Ikemen feature
		p2_value_empty_icon_offset = {0, 0}, --Ikemen feature
		p2_value_empty_icon_facing = 1, --Ikemen feature
		p2_value_empty_icon_scale = {1.0, 1.0}, --Ikemen feature
		p2_value_icon_spacing = {0, 0}, --Ikemen feature
		p2_value_snd = {-1, 0}, --Ikemen feature
		timer_offset = {0, 0}, --Ikemen feature
		timer_font = {-1, 0, 0, 255, 255, 255, -1}, --Ikemen feature
		timer_scale = {1.0, 1.0}, --Ikemen feature
		timer_text = "%i", --Ikemen feature
		timer_count = -1, --Ikemen feature
		timer_framespercount = 60, --Ikemen feature
		timer_displaytime = 10, --Ikemen feature
		stage_snd = {-1, 0}, --Ikemen feature
	},
	versusbgdef =
	{
		spr = '',
		bgclearcolor = {0, 0, 0},
	},
	demo_mode =
	{
		enabled = 1,
		select_enabled = 0, --not used in ikemen
		vsscreen_enabled = 0, --not used in ikemen
		fadein_time = 50, --Ikemen feature
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadein_anim = -1, --Ikemen feature
		fadeout_time = 50, --Ikemen feature
		fadeout_col = {0, 0, 0}, --Ikemen feature
		fadeout_anim = -1, --Ikemen feature
		title_waittime = 600,
		fight_endtime = 1500,
		fight_playbgm = 0,
		fight_stopbgm = 0,
		fight_bars_display = 0,
		intro_waitcycles = 1,
		debuginfo = 0,
	},
	continue_screen =
	{
		enabled = 1,
		sounds_enabled = 1, --Ikemen feature
		legacymode_enabled = 1, --Ikemen feature
		gameover_enabled = 1, --Ikemen feature
		fadein_time = 8, --Ikemen feature
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadein_anim = -1, --Ikemen feature
		fadeout_time = 120, --Ikemen feature
		fadeout_col = {0, 0, 0}, --Ikemen feature
		fadeout_anim = -1, --Ikemen feature
		pos = {160, 40},
		continue_text = 'Continue?',
		continue_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1},
		continue_scale = {1.0, 1.0},
		continue_offset = {0, 0},
		yes_text = 'Yes',
		yes_font = {'f-6x9.def', 0, 0, 191, 191, 191, -1},
		yes_scale = {1.0, 1.0},
		yes_offset = {-17, 20},
		yes_active_text = 'Yes',
		yes_active_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1},
		yes_active_scale = {1.0, 1.0},
		yes_active_offset = {-17, 20},
		no_text = 'No',
		no_font = {'f-6x9.def', 0, 0, 191, 191, 191, -1},
		no_scale = {1.0, 1.0},
		no_offset = {15, 20},
		no_active_text = 'No',
		no_active_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1},
		no_active_scale = {1.0, 1.0},
		no_active_offset = {15, 20},
		move_snd = {100, 0}, --Ikemen feature
		done_snd = {100, 1}, --Ikemen feature
		cancel_snd = {100, 2}, --Ikemen feature
		overlay_window = {0, 0, main.SP_Localcoord[1], main.SP_Localcoord[2]}, --Ikemen feature (0, 0, 320, 240)
		overlay_col = {0, 0, 0}, --Ikemen feature
		overlay_alpha = {0, 128}, --Ikemen feature
		p1_state = {5500, 5300}, --Ikemen feature
		p1_yes_state = {5510, 180}, --Ikemen feature
		p1_no_state = {5520, 170}, --Ikemen feature
		p2_state = {}, --Ikemen feature
		p2_yes_state = {}, --Ikemen feature
		p2_no_state = {}, --Ikemen feature
		p1_teammate_state = {}, --Ikemen feature
		p1_teammate_yes_state = {}, --Ikemen feature
		p1_teammate_no_state = {}, --Ikemen feature
		p2_teammate_state = {}, --Ikemen feature
		p2_teammate_yes_state = {}, --Ikemen feature
		p2_teammate_no_state = {}, --Ikemen feature
		credits_text = 'Credits: %i', --Ikemen feature
		credits_offset = {0, 0}, --Ikemen feature
		credits_font = {'jg.fnt', 0, 1, 255, 255, 255, -1}, --Ikemen feature
		credits_scale = {1.0, 1.0}, --Ikemen feature
		counter_endtime = 0, --Ikemen feature
		counter_starttime = 0, --Ikemen feature
		counter_anim = -1, --Ikemen feature
		counter_spr = {}, --Ikemen feature
		counter_offset = {0, 0}, --Ikemen feature
		counter_facing = 1, --Ikemen feature
		counter_scale = {1.0, 1.0}, --Ikemen feature
		counter_default_snd = {-1, 0}, --Ikemen feature
		counter_skipstart = 0, --Ikemen feature
		--counter_<num>_skiptime = 0, --Ikemen feature
		--counter_<num>_snd = {-1, 0}, --Ikemen feature
		counter_end_skiptime = 0, --Ikemen feature
		counter_end_snd = {-1, 0}, --Ikemen feature
	},
	continuebgdef =
	{
		spr = '', --Ikemen feature
	},
	game_over_screen =
	{
		enabled = 1,
		storyboard = '',
	},
	victory_screen =
	{
		enabled = 0,
		sounds_enabled = 0, --Ikemen feature
		cpu_enabled = 1, --Ikemen feature
		vs_enabled = 1, --Ikemen feature
		winner_teamko_enabled = 0, --Ikemen feature
		fadein_time = 0,
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadein_anim = -1, --Ikemen feature
		fadeout_time = 15,
		fadeout_col = {0, 0, 0}, --Ikemen feature
		fadeout_anim = -1, --Ikemen feature
		time = 300,
		p1_pos = {0, 0},
		p1_num = 1, --Ikemen feature
		p1_anim = -1, --Ikemen feature
		p1_spr = {9000, 2},
		p1_offset = {100, 20},
		p1_facing = 1,
		p1_scale = {1.0, 1.0},
		p1_window = {},
		p1_spacing = {0, 0}, --Ikemen feature
		p1_padding = 0, --Ikemen feature
		p1_name_offset = {0, 0},
		p1_name_font = {-1, 0, 1, 255, 255, 255, -1},
		p1_name_scale = {1.0, 1.0}, --Ikemen feature
		p2_pos = {0, 0}, --Ikemen feature
		p2_num = 0, --Ikemen feature
		p2_anim = -1, --Ikemen feature
		p2_spr = {9000, 2}, --Ikemen feature
		p2_offset = {100, 20}, --Ikemen feature
		p2_facing = 1, --Ikemen feature
		p2_scale = {1.0, 1.0}, --Ikemen feature
		p2_window = {}, --Ikemen feature
		p2_spacing = {0, 0}, --Ikemen feature
		p2_padding = 0, --Ikemen feature
		--p<pn>_member<num>_anim = -1, --Ikemen feature
		--p<pn>_member<num>_spr = {9000, 2}, --Ikemen feature
		--p<pn>_member<num>_offset = {0, 0}, --Ikemen feature
		--p<pn>_member<num>_scale = {1.0, 1.0}, --Ikemen feature
		--p<pn>_member<num>_slide_speed = {0, 0}, --Ikemen feature
		--p<pn>_member<num>_slide_dist = {0, 0}, --Ikemen feature
		p1_face2_anim = -1, --Ikemen feature
		p1_face2_spr = {}, --Ikemen feature
		p1_face2_offset = {0, 0}, --Ikemen feature
		p1_face2_facing = 1, --Ikemen feature
		p1_face2_scale = {1.0, 1.0}, --Ikemen feature
		p1_face2_window = {}, --Ikemen feature
		p2_face2_anim = -1, --Ikemen feature
		p2_face2_spr = {}, --Ikemen feature
		p2_face2_offset = {0, 0}, --Ikemen feature
		p2_face2_facing = -1, --Ikemen feature
		p2_face2_scale = {1.0, 1.0}, --Ikemen feature
		p2_face2_window = {}, --Ikemen feature
		p2_name_offset = {0, 0}, --Ikemen feature
		p2_name_font = {-1, 0, 1, 255, 255, 255, -1}, --Ikemen feature
		p2_name_scale = {1.0, 1.0}, --Ikemen feature
		winquote_text = 'Winner!',
		winquote_offset = {20, 192},
		winquote_spacing = {0, 0}, --Ikemen feature
		winquote_font = {-1, 0, 1, 255, 255, 255, -1},
		winquote_scale = {1.0, 1.0},
		winquote_delay = 2, --Ikemen feature
		winquote_displaytime = 0, --Ikemen feature
		winquote_textwrap = 'w',
		winquote_window = {},
		overlay_window = {0, 0, main.SP_Localcoord[1], main.SP_Localcoord[2]}, --Ikemen feature (0, 0, 320, 240)
		overlay_col = {0, 0, 0}, --Ikemen feature
		overlay_alpha = {0, 128}, --Ikemen feature
		p1_state = {}, --Ikemen feature
		p2_state = {}, --Ikemen feature
		p1_teammate_state = {}, --Ikemen feature
		p2_teammate_state = {}, --Ikemen feature
	},
	victorybgdef =
	{
		spr = '',
		bgclearcolor = {0, 0, 0},
	},
	win_screen =
	{
		enabled = 1,
		sounds_enabled = 1, --Ikemen feature
		fadein_time = 0,
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadein_anim = -1, --Ikemen feature
		fadeout_time = 64,
		fadeout_col = {0, 0, 0}, --Ikemen feature
		fadeout_anim = -1, --Ikemen feature
		pose_time = 300,
		wintext_text = 'Congratulations!',
		wintext_offset = {159, 70},
		wintext_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1},
		wintext_scale = {1.0, 1.0},
		wintext_displaytime = 0,
		wintext_layerno = 2,
		overlay_window = {0, 0, main.SP_Localcoord[1], main.SP_Localcoord[2]}, --Ikemen feature (0, 0, 320, 240)
		overlay_col = {0, 0, 0}, --Ikemen feature
		overlay_alpha = {0, 128}, --Ikemen feature
		p1_state = {180}, --Ikemen feature
		p2_state = {}, --Ikemen feature
		p1_teammate_state = {}, --Ikemen feature
		p2_teammate_state = {}, --Ikemen feature
	},
	winbgdef =
	{
		spr = '', --Ikemen feature
	},
	default_ending =
	{
		enabled = 0,
		storyboard = '',
	},
	end_credits =
	{
		enabled = 0,
		storyboard = '',
	},
	survival_results_screen =
	{
		enabled = 1,
		sounds_enabled = 1, --Ikemen feature
		fadein_time = 0,
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadein_anim = -1, --Ikemen feature
		fadeout_time = 64,
		fadeout_col = {0, 0, 0}, --Ikemen feature
		fadeout_anim = -1, --Ikemen feature
		show_time = 300,
		winstext_text = 'Rounds survived: %i',
		winstext_offset = {159, 70},
		winstext_font = {'jg.fnt', 0, 0, 255, 255, 255, -1},
		winstext_scale = {1.0, 1.0},
		winstext_displaytime = 0,
		winstext_layerno = 2,
		roundstowin = 5,
		overlay_window = {0, 0, main.SP_Localcoord[1], main.SP_Localcoord[2]}, --Ikemen feature (0, 0, 320, 240)
		overlay_col = {0, 0, 0}, --Ikemen feature
		overlay_alpha = {0, 128}, --Ikemen feature
		p1_state = {175, 170}, --Ikemen feature
		p1_win_state = {180}, --Ikemen feature
		p2_state = {}, --Ikemen feature
		p2_win_state = {}, --Ikemen feature
		p1_teammate_state = {}, --Ikemen feature
		p1_teammate_win_state = {}, --Ikemen feature
		p2_teammate_state = {}, --Ikemen feature
		p2_teammate_win_state = {}, --Ikemen feature
	},
	survivalresultsbgdef =
	{
		spr = '', --Ikemen feature
	},
	time_attack_results_screen =
	{
		enabled = 1, --Ikemen feature
		sounds_enabled = 1, --Ikemen feature
		fadein_time = 0, --Ikemen feature
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadein_anim = -1, --Ikemen feature
		fadeout_time = 64, --Ikemen feature
		fadeout_col = {0, 0, 0}, --Ikemen feature
		fadeout_anim = -1, --Ikemen feature
		show_time = 300, --Ikemen feature
		winstext_text = 'Clear Time: %m:%s.%x', --Ikemen feature
		winstext_offset = {159, 70}, --Ikemen feature
		winstext_font = {'jg.fnt', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		winstext_scale = {1.0, 1.0}, --Ikemen feature
		winstext_displaytime = 0, --Ikemen feature
		winstext_layerno = 2, --Ikemen feature
		overlay_window = {0, 0, main.SP_Localcoord[1], main.SP_Localcoord[2]}, --Ikemen feature (0, 0, 320, 240)
		overlay_col = {0, 0, 0}, --Ikemen feature
		overlay_alpha = {0, 128}, --Ikemen feature
		p1_state = {175, 170}, --Ikemen feature
		p1_win_state = {180}, --Ikemen feature
		p2_state = {}, --Ikemen feature
		p2_win_state = {}, --Ikemen feature
		p1_teammate_state = {}, --Ikemen feature
		p1_teammate_win_state = {}, --Ikemen feature
		p2_teammate_state = {}, --Ikemen feature
		p2_teammate_win_state = {}, --Ikemen feature
	},
	timeattackresultsbgdef =
	{
		spr = '', --Ikemen feature
	},
	resultsbgdef =
	{
		spr = '', --Ikemen feature
	},
	option_info =
	{
		fadein_time = 10,
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadein_anim = -1, --Ikemen feature
		fadeout_time = 10,
		fadeout_col = {0, 0, 0}, --Ikemen feature
		fadeout_anim = -1, --Ikemen feature
		title_offset = {159, 15},
		title_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1},
		title_scale = {1.0, 1.0},
		title_text = 'OPTIONS', --Ikemen feature
		menu_uselocalcoord = 0, --Ikemen feature
		menu_pos = {85, 33}, --Ikemen feature
		--menu_bg_<itemname>_anim = -1, --Ikemen feature
		--menu_bg_<itemname>_spr = {}, --Ikemen feature
		--menu_bg_<itemname>_offset = {0, 0}, --Ikemen feature
		--menu_bg_<itemname>_facing = 1, --Ikemen feature
		--menu_bg_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		--menu_bg_active_<itemname>_anim = -1, --Ikemen feature
		--menu_bg_active_<itemname>_spr = {}, --Ikemen feature
		--menu_bg_active_<itemname>_offset = {0, 0}, --Ikemen feature
		--menu_bg_active_<itemname>_facing = 1, --Ikemen feature
		--menu_bg_active_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_offset = {0, 0}, --Ikemen feature
		menu_item_font = {'f-6x9.def', 0, 1, 191, 191, 191, -1}, --Ikemen feature
		menu_item_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_active_offset = {0, 0}, --Ikemen feature
		menu_item_active_font = {'f-6x9.def', 0, 1, 255, 255, 255, -1}, --Ikemen feature
		menu_item_active_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_selected_offset = {0, 0}, --Ikemen feature
		menu_item_selected_font = {'f-6x9.def', 0, 1, 0, 247, 247, -1}, --Ikemen feature
		menu_item_selected_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_selected_active_offset = {0, 0}, --Ikemen feature
		menu_item_selected_active_font = {'f-6x9.def', 0, 1, 0, 247, 247, -1}, --Ikemen feature
		menu_item_selected_active_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_value_offset = {150, 0}, --Ikemen feature
		menu_item_value_font = {'f-6x9.def', 0, -1, 191, 191, 191, -1}, --Ikemen feature
		menu_item_value_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_value_active_offset = {150, 0}, --Ikemen feature
		menu_item_value_active_font = {'f-6x9.def', 0, -1, 255, 255, 255, -1}, --Ikemen feature
		menu_item_value_active_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_value_conflict_offset = {150, 0}, --Ikemen feature
		menu_item_value_conflict_font = {'f-6x9.def', 0, -1, 247, 0, 0, -1}, --Ikemen feature
		menu_item_value_conflict_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_info_offset = {150, 0}, --Ikemen feature
		menu_item_info_font = {'f-6x9.def', 0, -1, 247, 247, 0, -1}, --Ikemen feature
		menu_item_info_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_info_active_offset = {150, 0}, --Ikemen feature
		menu_item_info_active_font = {'f-6x9.def', 0, -1, 247, 247, 0, -1}, --Ikemen feature
		menu_item_info_active_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_spacing = {0, 14}, --Ikemen feature
		menu_window_margins_y = {0, 0}, --Ikemen feature
		menu_window_visibleitems = 13, --Ikemen feature
		menu_boxcursor_visible = 1, --Ikemen feature
		menu_boxcursor_coords = {-5, -10, 154, 3}, --Ikemen feature
		menu_boxcursor_col = {255, 255, 255}, --Ikemen feature
		menu_boxcursor_alpharange = {10, 40, 2, 255, 255, 0}, --Ikemen feature
		menu_boxbg_visible = 1, --Ikemen feature
		menu_boxbg_col = {0, 0, 0}, --Ikemen feature
		menu_boxbg_alpha = {0, 128}, --Ikemen feature
		menu_arrow_up_anim = -1, --Ikemen feature
		menu_arrow_up_spr = {}, --Ikemen feature
		menu_arrow_up_offset = {0, 0}, --Ikemen feature
		menu_arrow_up_facing = 1, --Ikemen feature
		menu_arrow_up_scale = {1.0, 1.0}, --Ikemen feature
		menu_arrow_down_anim = -1, --Ikemen feature
		menu_arrow_down_spr = {}, --Ikemen feature
		menu_arrow_down_offset = {0, 0}, --Ikemen feature
		menu_arrow_down_facing = 1, --Ikemen feature
		menu_arrow_down_scale = {1.0, 1.0}, --Ikemen feature
		menu_title_uppercase = 1, --Ikemen feature
		menu_valuename_none = 'None', --Ikemen feature
		menu_valuename_random = 'Random', --Ikemen feature
		menu_valuename_default = 'Default', --Ikemen feature
		menu_valuename_f = '(F%i)', --Ikemen feature
		menu_valuename_esc = '(Esc)', --Ikemen feature
		menu_valuename_page = '(Tab)', --Ikemen feature
		menu_valuename_nokey = 'Not used', --Ikemen feature
		menu_valuename_yes = 'Yes', --Ikemen feature
		menu_valuename_no = 'No', --Ikemen feature
		menu_valuename_enabled = 'Enabled', --Ikemen feature
		menu_valuename_disabled = 'Disabled', --Ikemen feature
		keymenu_p1_pos = {39, 33}, --Ikemen feature
		keymenu_p2_pos = {178, 33}, --Ikemen feature
		--keymenu_bg_<itemname>_anim = -1, --Ikemen feature
		--keymenu_bg_<itemname>_spr = {}, --Ikemen feature
		--keymenu_bg_<itemname>_offset = {0, 0}, --Ikemen feature
		--keymenu_bg_<itemname>_facing = 1, --Ikemen feature
		--keymenu_bg_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		--keymenu_bg_active_<itemname>_anim = -1, --Ikemen feature
		--keymenu_bg_active_<itemname>_spr = {}, --Ikemen feature
		--keymenu_bg_active_<itemname>_offset = {0, 0}, --Ikemen feature
		--keymenu_bg_active_<itemname>_facing = 1, --Ikemen feature
		--keymenu_bg_active_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		keymenu_item_p1_offset = {52, 0}, --Ikemen feature
		keymenu_item_p1_font = {'f-6x9.def', 0, 0, 0, 247, 247, -1}, --Ikemen feature
		keymenu_item_p1_scale = {1.0, 1.0}, --Ikemen feature
		keymenu_item_p2_offset = {52, 0}, --Ikemen feature
		keymenu_item_p2_font = {'f-6x9.def', 0, 0, 247, 0, 0, -1}, --Ikemen feature
		keymenu_item_p2_scale = {1.0, 1.0}, --Ikemen feature
		--unassigned 'keymenu.item' parameters use corresponding 'menu.item' values
		keymenu_item_spacing = {0, 12}, --Ikemen feature
		keymenu_item_value_offset = {101, 0}, --Ikemen feature
		keymenu_item_value_active_offset = {101, 0}, --Ikemen feature
		keymenu_item_value_conflict_offset = {101, 0}, --Ikemen feature
		keymenu_item_info_offset = {101, 0}, --Ikemen feature
		keymenu_item_info_active_offset = {101, 0}, --Ikemen feature
		keymenu_boxcursor_coords = {-5, -9, 106, 2}, --Ikemen feature
		keymenu_itemname_playerno = 'PLAYER %i', --Ikemen feature
		keymenu_itemname_configall = 'Config all', --Ikemen feature
		keymenu_itemname_up = 'Up', --Ikemen feature
		keymenu_itemname_down = 'Down', --Ikemen feature
		keymenu_itemname_left = 'Left', --Ikemen feature
		keymenu_itemname_right = 'Right', --Ikemen feature
		keymenu_itemname_a = 'A', --Ikemen feature
		keymenu_itemname_b = 'B', --Ikemen feature
		keymenu_itemname_c = 'C', --Ikemen feature
		keymenu_itemname_x = 'X', --Ikemen feature
		keymenu_itemname_y = 'Y', --Ikemen feature
		keymenu_itemname_z = 'Z', --Ikemen feature
		keymenu_itemname_start = 'Start', --Ikemen feature
		keymenu_itemname_d = 'D', --Ikemen feature
		keymenu_itemname_w = 'W', --Ikemen feature
		keymenu_itemname_menu = 'Menu', --Ikemen feature
		keymenu_itemname_back = 'Back', --Ikemen feature
		keymenu_itemname_page = 'Page', --Ikemen feature
		textinput_offset = {25, 32}, --Ikemen feature
		textinput_font = {'default-3x5.def', 0, 1, 191, 191, 191, -1}, --Ikemen feature
		textinput_scale = {1.0, 1.0}, --Ikemen feature
		textinput_port_text = 'Type in Host Port, e.g. 7500.\nPress ENTER to accept.\nPress ESC to cancel.', --Ikemen feature
		textinput_reswidth_text = 'Type in screen width.\nPress ENTER to accept.\nPress ESC to cancel.', --Ikemen feature
		textinput_resheight_text = 'Type in screen height.\nPress ENTER to accept.\nPress ESC to cancel.', --Ikemen feature
		textinput_overlay_window = {0, 0, main.SP_Localcoord[1], main.SP_Localcoord[2]}, --Ikemen feature (0, 0, 320, 240)
		textinput_overlay_col = {0, 0, 0}, --Ikemen feature
		textinput_overlay_alpha = {0, 128}, --Ikemen feature
		cursor_move_snd = {100, 0},
		cursor_done_snd = {100, 1},
		cancel_snd = {100, 2},
		--menu_itemname_difficulty = 'Difficulty Level', --Ikemen feature
		--menu_itemname_roundtime = 'Time Limit', --Ikemen feature
		--menu_itemname_lifemul = 'Life', --Ikemen feature
		--menu_itemname_singlevsteamlife = 'Single VS Team Life', --Ikemen feature
		--menu_itemname_gamespeed = 'Game FPS', --Ikemen feature
		--menu_itemname_roundsnumsingle = 'Rounds to Win (Single)', --Ikemen feature
		--menu_itemname_maxdrawgames = 'Max Draw Games', --Ikemen feature
		--menu_itemname_credits = 'Credits', --Ikemen feature
		--menu_itemname_aipalette = 'Arcade Palette', --Ikemen feature
		--menu_itemname_aisurvivalpalette = 'Survival Palette', --Ikemen feature
		--menu_itemname_airamping = 'AI Ramping', --Ikemen feature
		--menu_itemname_quickcontinue = 'Quick Continue', --Ikemen feature
		--menu_itemname_autoguard = 'Auto-Guard', --Ikemen feature
		--menu_itemname_stunbar = 'Dizzy', --Ikemen feature
		--menu_itemname_guardbar = 'Guard Break', --Ikemen feature
		--menu_itemname_redlifebar = 'Red Life', --Ikemen feature
		--menu_itemname_teamduplicates = 'Team Duplicates', --Ikemen feature
		--menu_itemname_teamlifeshare = 'Team Life Share', --Ikemen feature
		--menu_itemname_teampowershare = 'Team Power Share', --Ikemen feature
		--menu_itemname_roundsnumtag = 'Rounds to Win (Tag)', --Ikemen feature
		--menu_itemname_losekotag = 'Partner KOed Lose', --Ikemen feature
		--menu_itemname_mintag = 'Min Tag Chars', --Ikemen feature
		--menu_itemname_maxtag = 'Max Tag Chars', --Ikemen feature
		--menu_itemname_roundsnumsimul = 'Rounds to Win (Simul)', --Ikemen feature
		--menu_itemname_losekosimul = 'Player KOed Lose', --Ikemen feature
		--menu_itemname_minsimul = 'Min Simul Chars', --Ikemen feature
		--menu_itemname_maxsimul = 'Max Simul Chars', --Ikemen feature
		--menu_itemname_turnsrecoverybase = 'Turns Recovery Base', --Ikemen feature
		--menu_itemname_turnsrecoverybonus = 'Turns Recovery Bonus', --Ikemen feature
		--menu_itemname_minturns = 'Min Turns Chars', --Ikemen feature
		--menu_itemname_maxturns = 'Max Turns Chars', --Ikemen feature
		--menu_itemname_ratiorecoverybase = 'Ratio Recovery Base', --Ikemen feature
		--menu_itemname_ratiorecoverybonus = 'Ratio Recovery Bonus', --Ikemen feature
		--menu_itemname_ratio1life = 'Ratio 1 Life', --Ikemen feature
		--menu_itemname_ratio1attack = 'Ratio 1 Damage', --Ikemen feature
		--menu_itemname_ratio2life = 'Ratio 2 Life', --Ikemen feature
		--menu_itemname_ratio2attack = 'Ratio 2 Damage', --Ikemen feature
		--menu_itemname_ratio3life = 'Ratio 3 Life', --Ikemen feature
		--menu_itemname_ratio3attack = 'Ratio 3 Damage', --Ikemen feature
		--menu_itemname_ratio4life = 'Ratio 4 Life', --Ikemen feature
		--menu_itemname_ratio4attack = 'Ratio 4 Damage', --Ikemen feature
		--menu_itemname_resolution = 'Resolution', --Ikemen feature
		--menu_itemname_customres = 'Custom', --Ikemen feature
		--menu_itemname_fullscreen = 'Fullscreen', --Ikemen feature
		--menu_itemname_vretrace = 'VSync', --Ikemen feature
		--menu_itemname_msaa = 'MSAA', --Ikemen feature
		--menu_itemname_shaders = 'Shaders', --Ikemen feature
		--menu_itemname_noshader = 'Disable', --Ikemen feature
		--menu_itemname_mastervolume = 'Master Volume', --Ikemen feature
		--menu_itemname_bgmvolume = 'BGM Volume', --Ikemen feature
		--menu_itemname_sfxvolume = 'SFX Volume', --Ikemen feature
		--menu_itemname_audioducking = 'Audio Ducking', --Ikemen feature
		--menu_itemname_stereoeffects = "Stereo Effects", --Ikemen feature
		--menu_itemname_panningrange = "Panning Range", --Ikemen feature
		--menu_itemname_keyboard = 'Key Config', --Ikemen feature
		--menu_itemname_gamepad = 'Joystick Config', --Ikemen feature
		--menu_itemname_inputdefault = 'Default', --Ikemen feature
		--menu_itemname_players = 'Players', --Ikemen feature
		--menu_itemname_debugkeys = 'Debug Keys', --Ikemen feature
		--menu_itemname_debugmode = 'Debug Mode', --Ikemen feature
		--menu_itemname_helpermax = 'HelperMax', --Ikemen feature
		--menu_itemname_projectilemax = 'PlayerProjectileMax', --Ikemen feature
		--menu_itemname_explodmax = 'ExplodMax', --Ikemen feature
		--menu_itemname_afterimagemax = 'AfterImageMax', --Ikemen feature
		--menu_itemname_portchange = 'Port Change', --Ikemen feature
		--menu_itemname_default = 'Default Values', --Ikemen feature
		--menu_itemname_empty = '', --Ikemen feature
		--menu_itemname_back = 'Back', --Ikemen feature
		--menu_itemname_savereturn = 'Save and Return', --Ikemen feature
		--menu_itemname_return = 'Return Without Saving', --Ikemen feature
	},
	optionbgdef =
	{
		spr = '',
		bgclearcolor = {0, 0, 0},
	},
	replay_info =
	{
		fadein_time = 10, --Ikemen feature
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadein_anim = -1, --Ikemen feature
		fadeout_time = 10, --Ikemen feature
		fadeout_col = {0, 0, 0}, --Ikemen feature
		fadeout_anim = -1, --Ikemen feature
		title_offset = {159, 15}, --Ikemen feature
		title_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		title_scale = {1.0, 1.0}, --Ikemen feature
		title_text = 'REPLAY SELECT', --Ikemen feature
		menu_uselocalcoord = 0, --Ikemen feature
		menu_pos = {85, 33}, --Ikemen feature
		--menu_bg_<itemname>_anim = -1, --Ikemen feature
		--menu_bg_<itemname>_spr = {}, --Ikemen feature
		--menu_bg_<itemname>_offset = {0, 0}, --Ikemen feature
		--menu_bg_<itemname>_facing = 1, --Ikemen feature
		--menu_bg_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		--menu_bg_active_<itemname>_anim = -1, --Ikemen feature
		--menu_bg_active_<itemname>_spr = {}, --Ikemen feature
		--menu_bg_active_<itemname>_offset = {0, 0}, --Ikemen feature
		--menu_bg_active_<itemname>_facing = 1, --Ikemen feature
		--menu_bg_active_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_offset = {0, 0}, --Ikemen feature
		menu_item_font = {'f-6x9.def', 0, 1, 191, 191, 191, -1}, --Ikemen feature
		menu_item_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_active_offset = {0, 0}, --Ikemen feature
		menu_item_active_font = {'f-6x9.def', 0, 1, 255, 255, 255, -1}, --Ikemen feature
		menu_item_active_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_spacing = {0, 14}, --Ikemen feature
		menu_window_margins_y = {0, 0}, --Ikemen feature
		menu_window_visibleitems = 13, --Ikemen feature
		menu_boxcursor_visible = 1, --Ikemen feature
		menu_boxcursor_coords = {-5, -10, 154, 3}, --Ikemen feature
		menu_boxcursor_col = {255, 255, 255}, --Ikemen feature
		menu_boxcursor_alpharange = {10, 40, 2, 255, 255, 0}, --Ikemen feature
		menu_boxbg_visible = 1, --Ikemen feature
		menu_boxbg_col = {0, 0, 0}, --Ikemen feature
		menu_boxbg_alpha = {0, 128}, --Ikemen feature
		menu_arrow_up_anim = -1, --Ikemen feature
		menu_arrow_up_spr = {}, --Ikemen feature
		menu_arrow_up_offset = {0, 0}, --Ikemen feature
		menu_arrow_up_facing = 1, --Ikemen feature
		menu_arrow_up_scale = {1.0, 1.0}, --Ikemen feature
		menu_arrow_down_anim = -1, --Ikemen feature
		menu_arrow_down_spr = {}, --Ikemen feature
		menu_arrow_down_offset = {0, 0}, --Ikemen feature
		menu_arrow_down_facing = 1, --Ikemen feature
		menu_arrow_down_scale = {1.0, 1.0}, --Ikemen feature
		menu_title_uppercase = 1, --Ikemen feature
		cursor_move_snd = {100, 0}, --Ikemen feature
		cursor_done_snd = {100, 1}, --Ikemen feature
		cancel_snd = {100, 2}, --Ikemen feature
		menu_itemname_back = 'Back', --Ikemen feature
	},
	replaybgdef =
	{
		spr = '', --Ikemen feature
		bgclearcolor = {0, 0, 0}, --Ikemen feature
	},
	menu_info =
	{
		fadein_time = 0, --Ikemen feature
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadein_anim = -1, --Ikemen feature
		fadeout_time = 0, --Ikemen feature
		fadeout_col = {0, 0, 0}, --Ikemen feature
		fadeout_anim = -1, --Ikemen feature
		title_offset = {159, 15}, --Ikemen feature
		title_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		title_scale = {1.0, 1.0}, --Ikemen feature
		title_text = 'PAUSE', --Ikemen feature
		menu_uselocalcoord = 0, --Ikemen feature
		menu_pos = {85, 33}, --Ikemen feature
		--menu_bg_<itemname>_anim = -1, --Ikemen feature
		--menu_bg_<itemname>_spr = {}, --Ikemen feature
		--menu_bg_<itemname>_offset = {0, 0}, --Ikemen feature
		--menu_bg_<itemname>_facing = 1, --Ikemen feature
		--menu_bg_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		--menu_bg_active_<itemname>_anim = -1, --Ikemen feature
		--menu_bg_active_<itemname>_spr = {}, --Ikemen feature
		--menu_bg_active_<itemname>_offset = {0, 0}, --Ikemen feature
		--menu_bg_active_<itemname>_facing = 1, --Ikemen feature
		--menu_bg_active_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_offset = {0, 0}, --Ikemen feature
		menu_item_font = {'f-6x9.def', 0, 1, 191, 191, 191, -1}, --Ikemen feature
		menu_item_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_active_offset = {0, 0}, --Ikemen feature
		menu_item_active_font = {'f-6x9.def', 0, 1, 255, 255, 255, -1}, --Ikemen feature
		menu_item_active_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_selected_offset = {0, 0}, --Ikemen feature
		menu_item_selected_font = {'f-6x9.def', 0, 1, 0, 247, 247, -1}, --Ikemen feature
		menu_item_selected_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_selected_active_offset = {0, 0}, --Ikemen feature
		menu_item_selected_active_font = {'f-6x9.def', 0, 1, 0, 247, 247, -1}, --Ikemen feature
		menu_item_selected_active_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_value_offset = {150, 0}, --Ikemen feature
		menu_item_value_font = {'f-6x9.def', 0, -1, 191, 191, 191, -1}, --Ikemen feature
		menu_item_value_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_value_active_offset = {150, 0}, --Ikemen feature
		menu_item_value_active_font = {'f-6x9.def', 0, -1, 255, 255, 255, -1}, --Ikemen feature
		menu_item_value_active_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_spacing = {0, 14}, --Ikemen feature
		menu_window_margins_y = {0, 0}, --Ikemen feature
		menu_window_visibleitems = 13, --Ikemen feature
		menu_boxcursor_visible = 1, --Ikemen feature
		menu_boxcursor_coords = {-5, -10, 154, 3}, --Ikemen feature
		menu_boxcursor_col = {255, 255, 255}, --Ikemen feature
		menu_boxcursor_alpharange = {10, 40, 2, 255, 255, 0}, --Ikemen feature
		menu_boxbg_visible = 1, --Ikemen feature
		menu_boxbg_col = {0, 0, 0}, --Ikemen feature
		menu_boxbg_alpha = {0, 128}, --Ikemen feature
		menu_arrow_up_anim = -1, --Ikemen feature
		menu_arrow_up_spr = {}, --Ikemen feature
		menu_arrow_up_offset = {0, 0}, --Ikemen feature
		menu_arrow_up_facing = 1, --Ikemen feature
		menu_arrow_up_scale = {1.0, 1.0}, --Ikemen feature
		menu_arrow_down_anim = -1, --Ikemen feature
		menu_arrow_down_spr = {}, --Ikemen feature
		menu_arrow_down_offset = {0, 0}, --Ikemen feature
		menu_arrow_down_facing = 1, --Ikemen feature
		menu_arrow_down_scale = {1.0, 1.0}, --Ikemen feature
		menu_title_uppercase = 1, --Ikemen feature
		overlay_window = {0, 0, main.SP_Localcoord[1], main.SP_Localcoord[2]}, --Ikemen feature (0, 0, 320, 240)
		overlay_col = {0, 0, 0}, --Ikemen feature
		overlay_alpha = {0, 128}, --Ikemen feature
		cursor_move_snd = {100, 0}, --Ikemen feature
		cursor_done_snd = {100, 1}, --Ikemen feature
		cancel_snd = {100, 2}, --Ikemen feature
		enter_snd = {-1, 0}, --Ikemen feature
		movelist_pos = {10, 20}, --Ikemen feature
		movelist_title_offset = {150, 0}, --Ikemen feature
		movelist_title_font = {'Open_Sans.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		movelist_title_scale = {0.4, 0.4}, --Ikemen feature
		movelist_title_text = '%s', --Ikemen feature
		movelist_title_uppercase = 0, --Ikemen feature
		movelist_text_offset = {0, 12}, --Ikemen feature
		movelist_text_font = {'Open_Sans.def', 0, 1, 255, 255, 255, -1}, --Ikemen feature
		movelist_text_scale = {0.4, 0.4}, --Ikemen feature
		movelist_text_spacing = {1, 1}, --Ikemen feature
		movelist_text_text = 'Command List not found.', --Ikemen feature
		movelist_glyphs_offset = {0, 2}, --Ikemen feature
		movelist_glyphs_scale = {1.0, 1.0}, --Ikemen feature
		movelist_glyphs_spacing = {2, 0}, --Ikemen feature
		movelist_window_width = 300, --Ikemen feature
		movelist_window_margins_y = {20, 1}, --Ikemen feature
		movelist_window_visibleitems = 18, --Ikemen feature
		movelist_overlay_window = {0, 0, main.SP_Localcoord[1], main.SP_Localcoord[2]}, --Ikemen feature (0, 0, 320, 240)
		movelist_overlay_col = {0, 0, 0}, --Ikemen feature
		movelist_overlay_alpha = {0, 128}, --Ikemen feature
		movelist_arrow_up_anim = -1, --Ikemen feature
		movelist_arrow_up_spr = {}, --Ikemen feature
		movelist_arrow_up_offset = {0, 0}, --Ikemen feature
		movelist_arrow_up_facing = 1, --Ikemen feature
		movelist_arrow_up_scale = {1.0, 1.0}, --Ikemen feature
		movelist_arrow_down_anim = -1, --Ikemen feature
		movelist_arrow_down_spr = {}, --Ikemen feature
		movelist_arrow_down_offset = {0, 0}, --Ikemen feature
		movelist_arrow_down_facing = 1, --Ikemen feature
		movelist_arrow_down_scale = {1.0, 1.0}, --Ikemen feature
		--menu_itemname_back = 'Continue', --Ikemen feature
		--menu_itemname_keyboard = 'Key Config', --Ikemen feature
		--menu_itemname_gamepad = 'Joystick Config', --Ikemen feature
		--menu_itemname_inputdefault = 'Default', --Ikemen feature
		--menu_itemname_reset = 'Round Reset', --Ikemen feature
		--menu_itemname_reload = 'Rematch', --Ikemen feature
		--menu_itemname_commandlist = 'Command List', --Ikemen feature
		--menu_itemname_characterchange = 'Character Change', --Ikemen feature
		--menu_itemname_exit = 'Exit', --Ikemen feature
	},
	menubgdef =
	{
		spr = '', --Ikemen feature
		bgclearcolor = {0, 0, 0}, --Ikemen feature
	},
	training_info =
	{
		--same default values as menu_info
		--training specific parameters:
		menu_valuename_dummycontrol_cooperative = "Cooperative", --Ikemen feature
		menu_valuename_dummycontrol_ai = "AI", --Ikemen feature
		menu_valuename_dummycontrol_manual = "Manual", --Ikemen feature
		menu_valuename_ailevel_1 = "1", --Ikemen feature
		menu_valuename_ailevel_2 = "2", --Ikemen feature
		menu_valuename_ailevel_3 = "3", --Ikemen feature
		menu_valuename_ailevel_4 = "4", --Ikemen feature
		menu_valuename_ailevel_5 = "5", --Ikemen feature
		menu_valuename_ailevel_6 = "6", --Ikemen feature
		menu_valuename_ailevel_7 = "7", --Ikemen feature
		menu_valuename_ailevel_8 = "8", --Ikemen feature
		menu_valuename_dummymode_stand = "Stand", --Ikemen feature
		menu_valuename_dummymode_crouch = "Crouch", --Ikemen feature
		menu_valuename_dummymode_jump = "Jump", --Ikemen feature
		menu_valuename_dummymode_wjump = "W Jump", --Ikemen feature
		menu_valuename_guardmode_none = "None", --Ikemen feature
		menu_valuename_guardmode_auto = "Auto", --Ikemen feature
		menu_valuename_guardmode_all = "All", --Ikemen feature
		menu_valuename_guardmode_random = "Random", --Ikemen feature
		menu_valuename_fallrecovery_none = "None", --Ikemen feature
		menu_valuename_fallrecovery_ground = "Ground", --Ikemen feature
		menu_valuename_fallrecovery_air = "Air", --Ikemen feature
		menu_valuename_fallrecovery_random = "Random", --Ikemen feature
		menu_valuename_distance_any = "Any", --Ikemen feature
		menu_valuename_distance_close = "Close", --Ikemen feature
		menu_valuename_distance_medium = "Medium", --Ikemen feature
		menu_valuename_distance_far = "Far", --Ikemen feature
		menu_valuename_buttonjam_none = "None", --Ikemen feature
		menu_valuename_buttonjam_a = "A", --Ikemen feature
		menu_valuename_buttonjam_b = "B", --Ikemen feature
		menu_valuename_buttonjam_c = "C", --Ikemen feature
		menu_valuename_buttonjam_x = "X", --Ikemen feature
		menu_valuename_buttonjam_y = "Y", --Ikemen feature
		menu_valuename_buttonjam_z = "Z", --Ikemen feature
		menu_valuename_buttonjam_s = "Start", --Ikemen feature
		menu_valuename_buttonjam_d = "D", --Ikemen feature
		menu_valuename_buttonjam_w = "W", --Ikemen feature
		--menu_itemname_dummycontrol = "Dummy Control", --Ikemen feature
		--menu_itemname_ailevel = "AI Level", --Ikemen feature
		--menu_itemname_dummymode = "Dummy Mode", --Ikemen feature
		--menu_itemname_guardmode = "Guard Mode", --Ikemen feature
		--menu_itemname_fallrecovery = "Fall Recovery", --Ikemen feature
		--menu_itemname_distance = "Distance", --Ikemen feature
		--menu_itemname_buttonjam = "Button Jam", --Ikemen feature
	},
	trainingbgdef =
	{
		spr = '', --Ikemen feature
		bgclearcolor = {0, 0, 0}, --Ikemen feature
	},
	attract_mode =
	{
		enabled = 0, --Ikemen feature
		fadein_time = 10, --Ikemen feature
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadein_anim = -1, --Ikemen feature
		fadeout_time = 10, --Ikemen feature
		fadeout_col = {0, 0, 0}, --Ikemen feature
		fadeout_anim = -1, --Ikemen feature
		credits_key = '', --Ikemen feature
		options_key = 'F11', --Ikemen feature
		credits_snd = {-1, 0}, --Ikemen feature
		logo_storyboard = '', --Ikemen feature
		intro_storyboard = '', --Ikemen feature
		start_storyboard = '', --Ikemen feature
		start_time = 600, --Ikemen feature
		start_insert_text = 'Insert coin', --Ikemen feature
		start_insert_offset = {159, 185}, --Ikemen feature
		start_insert_font = {'jg.fnt', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		start_insert_scale = {1.0, 1.0}, --Ikemen feature
		start_insert_blinktime = 30, --Ikemen feature
		start_press_text = 'Press Start', --Ikemen feature
		start_press_offset = {159, 185}, --Ikemen feature
		start_press_font = {'jg.fnt', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		start_press_scale = {1.0, 1.0}, --Ikemen feature
		start_press_blinktime = 30, --Ikemen feature
		start_timer_offset = {310, 234}, --Ikemen feature
		start_timer_font = {'f-4x6.fnt', 0, -1, 255, 255, 255, -1}, --Ikemen feature
		start_timer_scale = {1.0, 1.0}, --Ikemen feature
		start_timer_text = '%i', --Ikemen feature
		start_timer_count = 60, --Ikemen feature
		start_timer_framespercount = 60, --Ikemen feature
		start_timer_displaytime = 10, --Ikemen feature
		start_done_snd = {100, 1}, --Ikemen feature
		credits_text = 'CREDITS: %2i', --Ikemen feature
		credits_offset = {159, 234}, --Ikemen feature
		credits_font = {'f-4x6.fnt', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		credits_scale = {1.0, 1.0}, --Ikemen feature
		title_offset = {159, 15}, --Ikemen feature
		title_font = {-1, 0, 0, 255, 255, 255, -1}, --Ikemen feature
		title_scale = {1.0, 1.0}, --Ikemen feature
		title_text = 'MAIN MENU', --Ikemen feature
		menu_next_key = '$D&$F', --Ikemen feature
		menu_previous_key = '$U&$B', --Ikemen feature
		menu_accept_key = 'a&b&c&x&y&z&s', --Ikemen feature
		menu_pos = {159, 158}, --Ikemen feature
		--menu_bg_<itemname>_anim = -1, --Ikemen feature
		--menu_bg_<itemname>_spr = {}, --Ikemen feature
		--menu_bg_<itemname>_offset = {0, 0}, --Ikemen feature
		--menu_bg_<itemname>_facing = 1, --Ikemen feature
		--menu_bg_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		--menu_bg_active_<itemname>_anim = -1, --Ikemen feature
		--menu_bg_active_<itemname>_spr = {}, --Ikemen feature
		--menu_bg_active_<itemname>_offset = {0, 0}, --Ikemen feature
		--menu_bg_active_<itemname>_facing = 1, --Ikemen feature
		--menu_bg_active_<itemname>_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_offset = {0, 0}, --Ikemen feature
		menu_item_font = {'f-6x9.def', 0, 0, 191, 191, 191, -1}, --Ikemen feature
		menu_item_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_active_offset = {0, 0}, --Ikemen feature
		menu_item_active_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		menu_item_active_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_spacing = {0, 13}, --Ikemen feature
		menu_window_margins_y = {12, 8}, --Ikemen feature
		menu_window_visibleitems = 5, --Ikemen feature
		menu_boxcursor_visible = 1, --Ikemen feature
		menu_boxcursor_coords = {-40, -10, 39, 2},
		menu_boxcursor_col = {255, 255, 255}, --Ikemen feature
		menu_boxcursor_alpharange = {10, 40, 2, 255, 255, 0}, --Ikemen feature
		menu_boxbg_visible = 0, --Ikemen feature
		menu_boxbg_col = {0, 0, 0}, --Ikemen feature
		menu_boxbg_alpha = {0, 128}, --Ikemen feature
		menu_arrow_up_anim = -1, --Ikemen feature
		menu_arrow_up_spr = {}, --Ikemen feature
		menu_arrow_up_offset = {0, 0}, --Ikemen feature
		menu_arrow_up_facing = 1, --Ikemen feature
		menu_arrow_up_scale = {1.0, 1.0}, --Ikemen feature
		menu_arrow_down_anim = -1, --Ikemen feature
		menu_arrow_down_spr = {}, --Ikemen feature
		menu_arrow_down_offset = {0, 0}, --Ikemen feature
		menu_arrow_down_facing = 1, --Ikemen feature
		menu_arrow_down_scale = {1.0, 1.0}, --Ikemen feature
		menu_title_uppercase = 1, --Ikemen feature
		cursor_move_snd = {100, 0}, --Ikemen feature
		cursor_done_snd = {100, 1}, --Ikemen feature
		cancel_snd = {100, 2}, --Ikemen feature
	},
	attractbgdef =
	{
		spr = '', --Ikemen feature
		bgclearcolor = {0, 0, 0}, --Ikemen feature
	},
	challenger_info =
	{
		enabled = 0, --Ikemen feature
		fadein_time = 0, --Ikemen feature
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadein_anim = -1, --Ikemen feature
		fadeout_time = 64, --Ikemen feature
		fadeout_col = {0, 0, 0}, --Ikemen feature
		fadeout_anim = -1, --Ikemen feature
		time = 0, --Ikemen feature
		pause_time = 0, --Ikemen feature
		snd_time = 0, --Ikemen feature
		snd = {-1, 0}, --Ikemen feature
		bg_anim = -1, --Ikemen feature
		bg_spr = {}, --Ikemen feature
		bg_offset = {0, 0}, --Ikemen feature
		bg_facing = 1, --Ikemen feature
		bg_scale = {1.0, 1.0}, --Ikemen feature
		bg_displaytime = 0, --Ikemen feature
		text = '', --Ikemen feature
		text_offset = {0, 0}, --Ikemen feature
		text_font = {-1, 0, 1, 255, 255, 255, -1}, --Ikemen feature
		text_scale = {1.0, 1.0}, --Ikemen feature
		text_displaytime = 0, --Ikemen feature
		text_layerno = 2, --Ikemen feature
		overlay_window = {0, 0, main.SP_Localcoord[1], main.SP_Localcoord[2]}, --Ikemen feature (0, 0, 320, 240)
		overlay_col = {0, 0, 0}, --Ikemen feature
		overlay_alpha = {0, 0}, --Ikemen feature
	},
	challengerbgdef =
	{
		spr = '', --Ikemen feature
	},
	dialogue_info =
	{
		enabled = 0, --Ikemen feature
		endtime = 0, --Ikemen feature
		switchtime = 0, --Ikemen feature
		skiptime = 0, --Ikemen feature
		skip_key = 'a', --Ikemen feature
		cancel_key = 'b&c&x&y&z&s', --Ikemen feature
		p1_bg_anim = -1, --Ikemen feature
		p1_bg_spr = {}, --Ikemen feature
		p1_bg_offset = {0, 0}, --Ikemen feature
		p1_bg_facing = 1, --Ikemen feature
		p1_bg_scale = {1.0, 1.0}, --Ikemen feature
		p2_bg_anim = -1, --Ikemen feature
		p2_bg_spr = {}, --Ikemen feature
		p2_bg_offset = {0, 0}, --Ikemen feature
		p2_bg_facing = 1, --Ikemen feature
		p2_bg_scale = {1.0, 1.0}, --Ikemen feature
		p1_face_spr = {9000, 0}, --Ikemen feature
		p1_face_offset = {0, 0}, --Ikemen feature
		p1_face_facing = 1, --Ikemen feature
		p1_face_scale = {1.0, 1.0}, --Ikemen feature
		p1_face_window = {}, --Ikemen feature
		p2_face_spr = {9000, 0}, --Ikemen feature
		p2_face_offset = {0, 0}, --Ikemen feature
		p2_face_facing = -1, --Ikemen feature
		p2_face_scale = {1.0, 1.0}, --Ikemen feature
		p2_face_window = {}, --Ikemen feature
		p1_name_offset = {0, 0}, --Ikemen feature
		p1_name_font = {-1, 0, 1, 255, 255, 255, -1}, --Ikemen feature
		p1_name_scale = {1.0, 1.0}, --Ikemen feature
		p2_name_offset = {0, 0}, --Ikemen feature
		p2_name_font = {-1, 0, 1, 255, 255, 255, -1}, --Ikemen feature
		p2_name_scale = {1.0, 1.0}, --Ikemen feature
		p1_text_offset = {20, 192}, --Ikemen feature
		p1_text_spacing = {0, 0}, --Ikemen feature
		p1_text_font = {'f-6x9.def', 0, 1, 255, 255, 255, -1}, --Ikemen feature
		p1_text_scale = {1.0, 1.0}, --Ikemen feature
		p1_text_delay = 2, --Ikemen feature
		p1_text_textwrap = 'w', --Ikemen feature
		p1_text_window = {}, --Ikemen feature
		p2_text_offset = {20, 192}, --Ikemen feature
		p2_text_spacing = {0, 0}, --Ikemen feature
		p2_text_font = {'f-6x9.def', 0, 1, 255, 255, 255, -1}, --Ikemen feature
		p2_text_scale = {1.0, 1.0}, --Ikemen feature
		p2_text_delay = 2, --Ikemen feature
		p2_text_textwrap = 'w', --Ikemen feature
		p2_text_window = {}, --Ikemen feature
		p1_active_anim = -1, --Ikemen feature
		p1_active_spr = {}, --Ikemen feature
		p1_active_offset = {0, 0}, --Ikemen feature
		p1_active_facing = 1, --Ikemen feature
		p1_active_scale = {1.0, 1.0}, --Ikemen feature
		p2_active_anim = -1, --Ikemen feature
		p2_active_spr = {}, --Ikemen feature
		p2_active_offset = {0, 0}, --Ikemen feature
		p2_active_facing = 1, --Ikemen feature
		p2_active_scale = {1.0, 1.0}, --Ikemen feature
	},
	hiscore_info =
	{
		enabled = 0, --Ikemen feature
		fadein_time = 50, --Ikemen feature
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadein_anim = -1, --Ikemen feature
		fadeout_time = 50, --Ikemen feature
		fadeout_col = {0, 0, 0}, --Ikemen feature
		fadeout_anim = -1, --Ikemen feature
		time = 360,
		pos = {0, 0}, --Ikemen feature
		title_offset = {0, 0}, --Ikemen feature
		title_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		title_scale = {1.0, 1.0}, --Ikemen feature
		title_text = 'Ranking %s', --Ikemen feature
		title_uppercase = 0, --Ikemen feature
		title_rank_offset = {0, 0}, --Ikemen feature
		title_rank_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		title_rank_scale = {1.0, 1.0}, --Ikemen feature
		title_rank_text = 'Rank', --Ikemen feature
		title_data_offset = {0, 0}, --Ikemen feature
		title_data_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		title_data_scale = {1.0, 1.0}, --Ikemen feature
		title_data_text = 'Result', --Ikemen feature
		title_name_offset = {0, 0}, --Ikemen feature
		title_name_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		title_name_scale = {1.0, 1.0}, --Ikemen feature
		title_name_text = 'Name', --Ikemen feature
		title_face_offset = {0, 0}, --Ikemen feature
		title_face_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		title_face_scale = {1.0, 1.0}, --Ikemen feature
		title_face_text = 'Character', --Ikemen feature
		item_offset = {0, 0}, --Ikemen feature
		item_spacing = {0, 0}, --Ikemen feature
		item_rank_offset = {0, 0}, --Ikemen feature
		item_rank_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		item_rank_scale = {1.0, 1.0}, --Ikemen feature
		item_rank_spacing = {0, 0}, --Ikemen feature
		item_rank_text = '%s', --Ikemen feature
		--item_rank_<num>_text = '%s', --Ikemen feature
		item_rank_active_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		item_rank_active_scale = {1.0, 1.0}, --Ikemen feature
		item_rank_active_switchtime = 3, --Ikemen feature
		item_rank_active2_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		item_rank_active2_scale = {1.0, 1.0}, --Ikemen feature
		item_data_offset = {0, 0}, --Ikemen feature
		item_data_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		item_data_scale = {1.0, 1.0}, --Ikemen feature
		item_data_spacing = {0, 0}, --Ikemen feature
		item_data_text = '%s', --Ikemen feature
		item_data_score_text = '%8s', --Ikemen feature
		item_data_time_text = "%m'%s''%x", --Ikemen feature
		item_data_win_text = 'Round %s', --Ikemen feature
		item_data_active_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		item_data_active_scale = {1.0, 1.0}, --Ikemen feature
		item_data_active_switchtime = 3, --Ikemen feature
		item_data_active2_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		item_data_active2_scale = {1.0, 1.0}, --Ikemen feature
		item_name_offset = {0, 0}, --Ikemen feature
		item_name_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		item_name_scale = {1.0, 1.0}, --Ikemen feature
		item_name_spacing = {0, 0}, --Ikemen feature
		item_name_text = '%3s', --Ikemen feature
		item_name_active_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		item_name_active_scale = {1.0, 1.0}, --Ikemen feature
		item_name_active_switchtime = 3, --Ikemen feature
		item_name_active2_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		item_name_active2_scale = {1.0, 1.0}, --Ikemen feature
		item_name_uppercase = 1, --Ikemen feature
		item_face_anim = -1, --Ikemen feature
		item_face_spr = {9000, 0}, --Ikemen feature
		item_face_offset = {0, 0}, --Ikemen feature
		item_face_facing = 1, --Ikemen feature
		item_face_scale = {1.0, 1.0}, --Ikemen feature
		item_face_window = {}, --Ikemen feature
		item_face_num = 1, --Ikemen feature
		item_face_spacing = {0, 0}, --Ikemen feature
		item_face_bg_anim = -1, --Ikemen feature
		item_face_bg_spr = {}, --Ikemen feature
		item_face_bg_offset = {0, 0}, --Ikemen feature
		item_face_bg_facing = 1, --Ikemen feature
		item_face_bg_scale = {1.0, 1.0}, --Ikemen feature
		item_face_unknown_anim = -1, --Ikemen feature
		item_face_unknown_spr = {}, --Ikemen feature
		item_face_unknown_offset = {0, 0}, --Ikemen feature
		item_face_unknown_facing = 1, --Ikemen feature
		item_face_unknown_scale = {1.0, 1.0}, --Ikemen feature
		timer_offset = {0, 0}, --Ikemen feature
		timer_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		timer_scale = {1.0, 1.0}, --Ikemen feature
		timer_text = '%s', --Ikemen feature
		timer_count = 99, --Ikemen feature
		timer_framespercount = 60, --Ikemen feature
		timer_displaytime = 10, --Ikemen feature
		window_width = 300, --Ikemen feature
		window_margins_y = {20, 1}, --Ikemen feature
		window_visibleitems = 10, --Ikemen feature
		overlay_window = {0, 0, main.SP_Localcoord[1], main.SP_Localcoord[2]}, --Ikemen feature (0, 0, 320, 240)
		overlay_col = {0, 0, 0}, --Ikemen feature
		overlay_alpha = {0, 128}, --Ikemen feature
		move_snd = {100, 0}, --Ikemen feature
		done_snd = {100, 1}, --Ikemen feature
		cancel_snd = {100, 2}, --Ikemen feature
		glyphs = {}, --Ikemen feature
	},
	hiscorebgdef =
	{
		spr = '', --Ikemen feature
		bgclearcolor = {0, 0, 0}, --Ikemen feature
	},
	warning_info =
	{
		title_offset = {159, 15}, --Ikemen feature
		title_font = {'f-6x9.def', 0, 0, 255, 255, 255, -1}, --Ikemen feature
		title_scale = {1.0, 1.0}, --Ikemen feature
		title_text = 'WARNING', --Ikemen feature
		text_offset = {25, 33}, --Ikemen feature
		text_font = {'f-6x9.def', 0, 1, 255, 255, 255, -1}, --Ikemen feature
		text_scale = {1.0, 1.0}, --Ikemen feature
		text_ratio_text = "Incorrect 'arcade.ratiomatches' settings detected.\nRefer to tutorial available in default select.def.", --Ikemen feature
		text_reload_text = 'Some selected options require Ikemen to be restarted.\nPress any key to exit the program.', --Ikemen feature
		text_noreload_text = 'Some selected options require Ikemen to be restarted.\nPress any key to continue.', --Ikemen feature
		text_keys_text = 'Conflict between button keys detected.\nAll keys should have unique assignment.\n\nPress any key to continue.\nPress ESC to reset.', --Ikemen feature
		text_pad_text = 'Controller not detected.\nCheck if your controller is plugged in.', --Ikemen feature
		text_shaders_text = 'No external OpenGL shaders detected.\nIkemen GO supports files with .vert and .frag extensions.\nShaders are loaded from "./external/shaders" directory.', --Ikemen feature
		overlay_window = {0, 0, main.SP_Localcoord[1], main.SP_Localcoord[2]}, --Ikemen feature (0, 0, 320, 240)
		overlay_col = {0, 0, 0}, --Ikemen feature
		overlay_alpha = {0, 128}, --Ikemen feature
		done_snd = {100, 0}, --Ikemen feature
		cancel_snd = {100, 2}, --Ikemen feature
	},
	glyphs =
	{
		['^A'] = {1, 0}, --A
		['^B'] = {2, 0}, --B
		['^C'] = {3, 0}, --C
		['^D'] = {4, 0}, --D
		['^W'] = {23, 0}, --W
		['^X'] = {24, 0}, --X
		['^Y'] = {25, 0}, --Y
		['^Z'] = {26, 0}, --Z
		['_+'] = {39, 0}, --+ (press at the same time as previous button)
		['_.'] = {40, 0}, --...
		['_DB'] = {41, 0}, --Down-Back
		['_D'] = {42, 0}, --Down
		['_DF'] = {43, 0}, --Down-Forward
		['_B'] = {44, 0}, --Back
		['_F'] = {46, 0}, --Forward
		['_UB'] = {47, 0}, --Up-Back
		['_U'] = {48, 0}, --Up
		['_UF'] = {49, 0}, --Up-Forward
		['^S'] = {51, 0}, --Start
		['^M'] = {52, 0}, --Menu (Select/Back)
		['^P'] = {53, 0}, --Any Punch (X / Y / Z)
		['^K'] = {54, 0}, --Any Kick (A / B / C)
		['^LP'] = {57, 0}, --Light Punch (X)
		['^MP'] = {58, 0}, --Medium Punch (Y)
		['^HP'] = {59, 0}, --Heavy Punch (Z)
		['^LK'] = {60, 0}, --Light Kick (A)
		['^MK'] = {61, 0}, --Medium Kick (B)
		['^HK'] = {62, 0}, --Heavy Kick (C)
		['^3K'] = {63, 0}, --3 Kick (A+B+C)
		['^3P'] = {64, 0}, --3 Punch (X+Y+Z)
		['^2K'] = {65, 0}, --2 Kick (A+B / B+C / A+C)
		['^2P'] = {66, 0}, --2 Punch (X+Y / Y+Z / X+Z)
		['_-'] = {90, 0}, --Arrow (tap following Button immediately - use in combos)
		['_!'] = {91, 0}, --Continue Arrow (follow with this move)
		['~DB'] = {92, 0}, --hold Down-Back
		['~D'] = {93, 0}, --hold Down
		['~DF'] = {94, 0}, --hold Down-Forward
		['~B'] = {95, 0}, --hold Back
		['~F'] = {96, 0}, --hold Forward
		['~UB'] = {97, 0}, --hold Up-Back
		['~U'] = {98, 0}, --hold Up
		['~UF'] = {99, 0}, --hold Up-Forward
		['_HCB'] = {100, 0}, --1/2 Circle Back
		['_HUF'] = {101, 0}, --1/2 Circle Forward Up
		['_HCF'] = {102, 0}, --1/2 Circle Forward
		['_HUB'] = {103, 0}, --1/2 Circle Back Up
		['_QFD'] = {104, 0}, --1/4 Circle Forward Down
		['_QDB'] = {105, 0}, --1/4 Circle Down Back (QCB/QDB)
		['_QCB'] = {105, 0}, --1/4 Circle Down Back (QCB/QDB)
		['_QBU'] = {106, 0}, --1/4 Circle Back Up
		['_QUF'] = {107, 0}, --1/4 Circle Up Forward
		['_QBD'] = {108, 0}, --1/4 Circle Back Down
		['_QDF'] = {109, 0}, --1/4 Circle Down Forward (QCF/QDF)
		['_QCF'] = {109, 0}, --1/4 Circle Down Forward (QCF/QDF)
		['_QFU'] = {110, 0}, --1/4 Circle Forward Up
		['_QUB'] = {111, 0}, --1/4 Circle Up Back
		['_FDF'] = {112, 0}, --Full Clock Forward
		['_FUB'] = {113, 0}, --Full Clock Back
		['_FUF'] = {114, 0}, --Full Count Forward
		['_FDB'] = {115, 0}, --Full Count Back
		['_XFF'] = {116, 0}, --2x Forward
		['_XBB'] = {117, 0}, --2x Back
		['_DSF'] = {118, 0}, --Dragon Screw Forward
		['_DSB'] = {119, 0}, --Dragon Screw Back
		['_AIR'] = {121, 0}, --AIR
		['_TAP'] = {122, 0}, --TAP
		['_MAX'] = {123, 0}, --MAX
		['_EX'] = {124, 0}, --EX
		['_^'] = {127, 0}, --Air
		['_='] = {128, 0}, --Squatting
		['_)'] = {129, 0}, --Close
		['_('] = {130, 0}, --Away
		['_`'] = {135, 0}, --Small Dot
	},
	anim = {},
}

function motif.setBaseTitleInfo()
	motif.title_info.menu_itemname_arcade = "ARCADE"
	motif.title_info.menu_itemname_versus = "VS MODE"
	motif.title_info.menu_itemname_teamarcade = "TEAM ARCADE"
	motif.title_info.menu_itemname_teamversus = "TEAM VS"
	motif.title_info.menu_itemname_teamcoop = "TEAM CO-OP"
	motif.title_info.menu_itemname_survival = "SURVIVAL"
	motif.title_info.menu_itemname_survivalcoop = "SURVIVAL CO-OP"
	motif.title_info.menu_itemname_training = "TRAINING"
	motif.title_info.menu_itemname_watch = "WATCH"
	motif.title_info.menu_itemname_options = "OPTIONS"
	motif.title_info.menu_itemname_exit = "EXIT"
	if main.t_sort.title_info == nil then
		main.t_sort.title_info = {}
	end
	main.t_sort.title_info.menu = {
		"arcade",
		"versus",
		"teamarcade",
		"teamversus",
		"teamcoop",
		"survival",
		"survivalcoop",
		"training",
		"watch",
		"options",
		"exit",
	}
	hook.run("motif.setBaseTitleInfo")
end

function motif.setBaseOptionInfo()
	motif.option_info.menu_itemname_menugame = "Game Settings"
	motif.option_info.menu_itemname_menugame_difficulty = "Difficulty Level"
	motif.option_info.menu_itemname_menugame_roundtime = "Time Limit"
	motif.option_info.menu_itemname_menugame_lifemul = "Life"
	motif.option_info.menu_itemname_menugame_singlevsteamlife = "Single VS Team Life"
	motif.option_info.menu_itemname_menugame_gamespeed = "Game FPS"
	motif.option_info.menu_itemname_menugame_roundsnumsingle = "Rounds to Win (Single)"
	motif.option_info.menu_itemname_menugame_maxdrawgames = "Max Draw Games"
	motif.option_info.menu_itemname_menugame_credits = "Credits"
	motif.option_info.menu_itemname_menugame_aipalette = "Arcade Palette"
	motif.option_info.menu_itemname_menugame_aisurvivalpalette = "Survival Palette"
	motif.option_info.menu_itemname_menugame_airamping = "AI Ramping"
	motif.option_info.menu_itemname_menugame_quickcontinue = "Quick Continue"
	motif.option_info.menu_itemname_menugame_autoguard = "Auto-Guard"
	motif.option_info.menu_itemname_menugame_stunbar = "Dizzy"
	motif.option_info.menu_itemname_menugame_guardbar = "Guard Break"
	motif.option_info.menu_itemname_menugame_redlifebar = "Red Life"
	motif.option_info.menu_itemname_menugame_teamduplicates = "Team Duplicates"
	motif.option_info.menu_itemname_menugame_teamlifeshare = "Team Life Share"
	motif.option_info.menu_itemname_menugame_teampowershare = "Team Power Share"
	motif.option_info.menu_itemname_menugame_empty = ""
	motif.option_info.menu_itemname_menugame_menutag = "Tag Settings"
	motif.option_info.menu_itemname_menugame_menutag_roundsnumtag = "Rounds to Win (Tag)"
	motif.option_info.menu_itemname_menugame_menutag_losekotag = "Partner KOed Lose"
	motif.option_info.menu_itemname_menugame_menutag_empty = ""
	motif.option_info.menu_itemname_menugame_menutag_mintag = "Min Tag Chars"
	motif.option_info.menu_itemname_menugame_menutag_maxtag = "Max Tag Chars"
	motif.option_info.menu_itemname_menugame_menutag_empty = ""
	motif.option_info.menu_itemname_menugame_menutag_back = "Back"
	motif.option_info.menu_itemname_menugame_menusimul = "Simul Settings"
	motif.option_info.menu_itemname_menugame_menusimul_roundsnumsimul = "Rounds to Win (Simul)"
	motif.option_info.menu_itemname_menugame_menusimul_losekosimul = "Player KOed Lose"
	motif.option_info.menu_itemname_menugame_menusimul_empty = ""
	motif.option_info.menu_itemname_menugame_menusimul_minsimul = "Min Simul Chars"
	motif.option_info.menu_itemname_menugame_menusimul_maxsimul = "Max Simul Chars"
	motif.option_info.menu_itemname_menugame_menusimul_empty = ""
	motif.option_info.menu_itemname_menugame_menusimul_back = "Back"
	motif.option_info.menu_itemname_menugame_menuturns = "Turns Settings"
	motif.option_info.menu_itemname_menugame_menuturns_turnsrecoverybase = "Turns Recovery Base"
	motif.option_info.menu_itemname_menugame_menuturns_turnsrecoverybonus = "Turns Recovery Bonus"
	motif.option_info.menu_itemname_menugame_menuturns_empty = ""
	motif.option_info.menu_itemname_menugame_menuturns_minturns = "Min Turns Chars"
	motif.option_info.menu_itemname_menugame_menuturns_maxturns = "Max Turns Chars"
	motif.option_info.menu_itemname_menugame_menuturns_empty = ""
	motif.option_info.menu_itemname_menugame_menuturns_back = "Back"
	motif.option_info.menu_itemname_menugame_menuratio = "Ratio Settings"
	motif.option_info.menu_itemname_menugame_menuratio_ratiorecoverybase = "Ratio Recovery Base"
	motif.option_info.menu_itemname_menugame_menuratio_ratiorecoverybonus = "Ratio Recovery Bonus"
	motif.option_info.menu_itemname_menugame_menuratio_empty = ""
	motif.option_info.menu_itemname_menugame_menuratio_ratio1life = "Ratio 1 Life"
	motif.option_info.menu_itemname_menugame_menuratio_ratio1attack = "Ratio 1 Damage"
	motif.option_info.menu_itemname_menugame_menuratio_ratio2life = "Ratio 2 Life"
	motif.option_info.menu_itemname_menugame_menuratio_ratio2attack = "Ratio 2 Damage"
	motif.option_info.menu_itemname_menugame_menuratio_ratio3life = "Ratio 3 Life"
	motif.option_info.menu_itemname_menugame_menuratio_ratio3attack = "Ratio 3 Damage"
	motif.option_info.menu_itemname_menugame_menuratio_ratio4life = "Ratio 4 Life"
	motif.option_info.menu_itemname_menugame_menuratio_ratio4attack = "Ratio 4 Damage"
	motif.option_info.menu_itemname_menugame_menuratio_empty = ""
	motif.option_info.menu_itemname_menugame_menuratio_back = "Back"
	motif.option_info.menu_itemname_menugame_back = "Back"

	motif.option_info.menu_itemname_menuvideo = "Video Settings"
	motif.option_info.menu_itemname_menuvideo_resolution = "Resolution" --reserved submenu
	-- Resolution is assigned based on values used in itemname suffix (e.g. 320x240)
	motif.option_info.menu_itemname_menuvideo_resolution_320x240 = "320x240    (4:3 QVGA)"
	motif.option_info.menu_itemname_menuvideo_resolution_640x480 = "640x480    (4:3 VGA)"
	motif.option_info.menu_itemname_menuvideo_resolution_960x720 = "960x720    (4:3 HD)"
	motif.option_info.menu_itemname_menuvideo_resolution_1280x720 = "1280x720   (16:9 HD)"
	motif.option_info.menu_itemname_menuvideo_resolution_1600x900 = "1600x900   (16:9 HD+)"
	motif.option_info.menu_itemname_menuvideo_resolution_1920x1080 = "1920x1080  (16:9 FHD)"
	motif.option_info.menu_itemname_menuvideo_resolution_empty = ""
	motif.option_info.menu_itemname_menuvideo_resolution_customres = "Custom"
	motif.option_info.menu_itemname_menuvideo_resolution_back = "Back"
	motif.option_info.menu_itemname_menuvideo_fullscreen = "Fullscreen"
	motif.option_info.menu_itemname_menuvideo_vretrace = "VSync"
	motif.option_info.menu_itemname_menuvideo_msaa = "MSAA"
	motif.option_info.menu_itemname_menuvideo_shaders = "Shaders" --reserved submenu
	-- This list is populated with shaders existing in 'external/shaders' directory
	motif.option_info.menu_itemname_menuvideo_shaders_empty = ""
	motif.option_info.menu_itemname_menuvideo_shaders_noshader = "Disable"
	motif.option_info.menu_itemname_menuvideo_shaders_back = "Back"
	motif.option_info.menu_itemname_menuvideo_empty = ""
	motif.option_info.menu_itemname_menuvideo_back = "Back"

	motif.option_info.menu_itemname_menuaudio = "Audio Settings"
	motif.option_info.menu_itemname_menuaudio_mastervolume = "Master Volume"
	motif.option_info.menu_itemname_menuaudio_bgmvolume = "BGM Volume"
	motif.option_info.menu_itemname_menuaudio_sfxvolume = "SFX Volume"
	motif.option_info.menu_itemname_menuaudio_audioducking = "Audio Ducking"
	motif.option_info.menu_itemname_menuaudio_stereoeffects = "Stereo Effects"
	motif.option_info.menu_itemname_menuaudio_panningrange = "Panning Range"
	motif.option_info.menu_itemname_menuaudio_empty = ""
	motif.option_info.menu_itemname_menuaudio_back = "Back"

	motif.option_info.menu_itemname_menuinput = "Input Settings"
	motif.option_info.menu_itemname_menuinput_keyboard = "Key Config"
	motif.option_info.menu_itemname_menuinput_gamepad = "Joystick Config"
	motif.option_info.menu_itemname_menuinput_empty = ""
	motif.option_info.menu_itemname_menuinput_inputdefault = "Default"
	motif.option_info.menu_itemname_menuinput_back = "Back"

	motif.option_info.menu_itemname_menuengine = "Engine Settings"
	motif.option_info.menu_itemname_menuengine_players = "Players"
	motif.option_info.menu_itemname_menuengine_debugkeys = "Debug Keys"
	motif.option_info.menu_itemname_menuengine_debugmode = "Debug Mode"
	motif.option_info.menu_itemname_menuengine_empty = ""
	motif.option_info.menu_itemname_menuengine_helpermax = "HelperMax"
	motif.option_info.menu_itemname_menuengine_projectilemax = "PlayerProjectileMax"
	motif.option_info.menu_itemname_menuengine_explodmax = "ExplodMax"
	motif.option_info.menu_itemname_menuengine_afterimagemax = "AfterImageMax"
	motif.option_info.menu_itemname_menuengine_empty = ""
	motif.option_info.menu_itemname_menuengine_back = "Back"

	motif.option_info.menu_itemname_empty = ""
	motif.option_info.menu_itemname_portchange = "Port Change"
	motif.option_info.menu_itemname_default = "Default Values"
	motif.option_info.menu_itemname_empty = ""
	motif.option_info.menu_itemname_savereturn = "Save and Return"
	motif.option_info.menu_itemname_return = "Return Without Saving"
	-- Default options screen order.
	if main.t_sort.option_info == nil then
		main.t_sort.option_info = {}
	end
	main.t_sort.option_info.menu = {
		"menugame",
		"menugame_difficulty",
		"menugame_roundtime",
		"menugame_lifemul",
		"menugame_singlevsteamlife",
		"menugame_gamespeed",
		"menugame_roundsnumsingle",
		"menugame_maxdrawgames",
		"menugame_credits",
		"menugame_aipalette",
		"menugame_aisurvivalpalette",
		"menugame_airamping",
		"menugame_quickcontinue",
		"menugame_autoguard",
		"menugame_stunbar",
		"menugame_guardbar",
		"menugame_redlifebar",
		"menugame_teamduplicates",
		"menugame_teamlifeshare",
		"menugame_teampowershare",
		"menugame_empty",
		"menugame_menutag",
		"menugame_menutag_roundsnumtag",
		"menugame_menutag_losekotag",
		"menugame_menutag_empty",
		"menugame_menutag_mintag",
		"menugame_menutag_maxtag",
		"menugame_menutag_empty",
		"menugame_menutag_back",
		"menugame_menusimul",
		"menugame_menusimul_roundsnumsimul",
		"menugame_menusimul_losekosimul",
		"menugame_menusimul_empty",
		"menugame_menusimul_minsimul",
		"menugame_menusimul_maxsimul",
		"menugame_menusimul_empty",
		"menugame_menusimul_back",
		"menugame_menuturns",
		"menugame_menuturns_turnsrecoverybase",
		"menugame_menuturns_turnsrecoverybonus",
		"menugame_menuturns_empty",
		"menugame_menuturns_minturns",
		"menugame_menuturns_maxturns",
		"menugame_menuturns_empty",
		"menugame_menuturns_back",
		"menugame_menuratio",
		"menugame_menuratio_ratiorecoverybase",
		"menugame_menuratio_ratiorecoverybonus",
		"menugame_menuratio_empty",
		"menugame_menuratio_ratio1life",
		"menugame_menuratio_ratio1attack",
		"menugame_menuratio_ratio2life",
		"menugame_menuratio_ratio2attack",
		"menugame_menuratio_ratio3life",
		"menugame_menuratio_ratio3attack",
		"menugame_menuratio_ratio4life",
		"menugame_menuratio_ratio4attack",
		"menugame_menuratio_empty",
		"menugame_menuratio_back",
		"menugame_back",
		"menuvideo",
		"menuvideo_resolution",
		"menuvideo_resolution_320x240",
		"menuvideo_resolution_640x480",
		"menuvideo_resolution_960x720",
		"menuvideo_resolution_1280x720",
		"menuvideo_resolution_1600x900",
		"menuvideo_resolution_1920x1080",
		"menuvideo_resolution_empty",
		"menuvideo_resolution_customres",
		"menuvideo_resolution_back",
		"menuvideo_fullscreen",
		"menuvideo_vretrace",
		"menuvideo_msaa",
		"menuvideo_shaders",
		"menuvideo_shaders_empty",
		"menuvideo_shaders_noshader",
		"menuvideo_shaders_back",
		"menuvideo_empty",
		"menuvideo_back",
		"menuaudio",
		"menuaudio_mastervolume",
		"menuaudio_bgmvolume",
		"menuaudio_sfxvolume",
		"menuaudio_audioducking",
		"menuaudio_stereoeffects",
		"menuaudio_panningrange",
		"menuaudio_empty",
		"menuaudio_back",
		"menuinput",
		"menuinput_keyboard",
		"menuinput_gamepad",
		"menuinput_empty",
		"menuinput_inputdefault",
		"menuinput_back",
		"menuengine",
		"menuengine_players",
		"menuengine_debugkeys",
		"menuengine_debugmode",
		"menuengine_empty",
		"menuengine_helpermax",
		"menuengine_projectilemax",
		"menuengine_explodmax",
		"menuengine_afterimagemax",
		"menuengine_empty",
		"menuengine_back",
		"empty",
		"portchange",
		"default",
		"empty",
		"savereturn",
		"return",
	}
	hook.run("motif.setBaseOptionInfo")
end

function motif.setBaseMenuInfo()
	motif.menu_info.menu_itemname_back = "Continue"
	motif.menu_info.menu_itemname_menuinput = "Button Config"
	motif.menu_info.menu_itemname_menuinput_keyboard = "Key Config"
	motif.menu_info.menu_itemname_menuinput_gamepad = "Joystick Config"
	motif.menu_info.menu_itemname_menuinput_empty = ""
	motif.menu_info.menu_itemname_menuinput_inputdefault = "Default"
	motif.menu_info.menu_itemname_menuinput_back = "Back"
	--menu_itemname_reset = "Round Reset"
	--menu_itemname_reload = "Rematch"
	motif.menu_info.menu_itemname_commandlist = "Command List"
	motif.menu_info.menu_itemname_characterchange = "Character Change"
	motif.menu_info.menu_itemname_exit = "Exit"
	if main.t_sort.menu_info == nil then
		main.t_sort.menu_info = {}
	end
	main.t_sort.menu_info.menu = {
		"back",
		"menuinput",
		"menuinput_keyboard",
		"menuinput_gamepad",
		"menuinput_empty",
		"menuinput_inputdefault",
		"menuinput_back",
		--"reset",
		--"reload",
		"commandlist",
		"characterchange",
		"exit",
	}
	hook.run("motif.setBaseMenuInfo")
end

function motif.setBaseTrainingInfo()
	motif.training_info.menu_itemname_back = "Continue"
	motif.training_info.menu_itemname_menutraining = "Training Menu"
	motif.training_info.menu_itemname_menutraining_dummycontrol = "Dummy Control"
	motif.training_info.menu_itemname_menutraining_ailevel = "AI Level"
	motif.training_info.menu_itemname_menutraining_dummymode = "Dummy Mode"
	motif.training_info.menu_itemname_menutraining_guardmode = "Guard Mode"
	motif.training_info.menu_itemname_menutraining_fallrecovery = "Fall Recovery"
	motif.training_info.menu_itemname_menutraining_distance = "Distance"
	motif.training_info.menu_itemname_menutraining_buttonjam = "Button Jam"
	motif.training_info.menu_itemname_menutraining_back = "Back"
	motif.training_info.menu_itemname_menuinput = "Button Config"
	motif.training_info.menu_itemname_menuinput_keyboard = "Key Config"
	motif.training_info.menu_itemname_menuinput_gamepad = "Joystick Config"
	motif.training_info.menu_itemname_menuinput_empty = ""
	motif.training_info.menu_itemname_menuinput_inputdefault = "Default"
	motif.training_info.menu_itemname_menuinput_back = "Back"
	--motif.training_info.menu_itemname_reset = "Round Reset"
	--motif.training_info.menu_itemname_reload = "Rematch"
	motif.training_info.menu_itemname_commandlist = "Command List"
	motif.training_info.menu_itemname_characterchange = "Character Change"
	motif.training_info.menu_itemname_exit = "Exit"
	if main.t_sort.training_info == nil then
		main.t_sort.training_info = {}
	end
	main.t_sort.training_info.menu = {
		"back",
		"menutraining",
		"menutraining_dummycontrol",
		"menutraining_ailevel",
		"menutraining_dummymode",
		"menutraining_guardmode",
		"menutraining_fallrecovery",
		"menutraining_distance",
		"menutraining_buttonjam",
		"menutraining_back",
		"menuinput",
		"menuinput_keyboard",
		"menuinput_gamepad",
		"menuinput_empty",
		"menuinput_inputdefault",
		"menuinput_back",
		--"reset",
		--"reload",
		"commandlist",
		"characterchange",
		"exit",
	}
	hook.run("motif.setBaseTrainingInfo")
end

--;===========================================================
--; PARSE SCREENPACK
--;===========================================================
--here starts proper screenpack DEF file parsing
main.t_fntDefault = {0, 0, 255, 255, 255, -1} -- bank, align, r, g, b, ttf height
main.t_sort = {}
local t = {}
local pos = t
local pos_sort = main.t_sort
local def_pos = motif
t.anim = {}
t.fileDir = main.motifDir
t.fileName = main.motifFile
local tmp = ''
local group = ''
--local file = io.open(motif.def, 'r')
--for line in file:lines() do
for line in main.motifData:gmatch('([^\n]*)\n?') do
	line = line:gsub('%s*;.*$', '')
	if line:match('^[^%g]*%s*%[.-%s*%]%s*$') then --matched [] group
		line = line:match('%[(.-)%s*%]%s*$') --match text between []
		line = line:gsub('[%. ]', '_') --change . and space to _
		group = tostring(line:lower())
		if group:match('infobox_text$') then
			t[group] = ''
		elseif group:match('^begin_action_[0-9]+$') then --matched anim
			group = tonumber(group:match('^begin_action_([0-9]+)$'))
			t.anim[group] = {}
			pos = t.anim[group]
		else --matched other []
			t[group] = {}
			main.t_sort[group] = {}
			pos = t[group]
			pos_sort = main.t_sort[group]
			def_pos = motif[group]
		end
	elseif type(t[group]) == 'string' then
		if t[group] == '' then
			t[group] = line
		else
			t[group] = t[group] .. '\n' .. line
		end
	else --matched non [] line
		local param, value = line:match('^%s*([^=]-)%s*=%s*(.-)%s*$')
		if param ~= nil then
			param = param:gsub('[%. ]', '_') --change param . and space to _
			if group ~= 'glyphs' then
				param = param:lower() --lowercase param
			end
			if value ~= nil and def_pos ~= nil then --let's check if it's even a valid param
				if value == '' and (type(def_pos[param]) == 'number' or type(def_pos[param]) == 'table') then --text should remain empty
					value = nil
				end
			end
		end
		if param ~= nil and value ~= nil then --param = value pattern matched
			value = value:gsub('"', '') --remove brackets from value
			value = value:gsub('^(%.[0-9])', '0%1') --add 0 before dot if missing at the beginning of matched string
			value = value:gsub('([^0-9])(%.[0-9])', '%10%2') --add 0 before dot if missing anywhere else
			value = value:gsub(',%s*$', '') --remove dummy ','
			if param:match('^font[0-9]+') then --font declaration param matched
				if pos.font == nil then
					pos.font = {}
					pos.font_height = {}
				end
				local num = tonumber(param:match('font([0-9]+)'))
				if param:match('_height$') then
					pos.font_height[num] = main.f_dataType(value)
				else
					value = value:gsub('\\', '/')
					pos.font[num] = tostring(value)
				end
			elseif pos[param] == nil or param:match('_itemname_') then --mugen takes into account only first occurrence
				if param:match('_font$') then --assign default font values if needed (also ensure that there are multiple values in the first place)
					local _, n = value:gsub(',', '')
					for i = n + 1, #main.t_fntDefault do
						value = value:gsub(',?%s*$', ',' .. main.t_fntDefault[i])
					end
				end
				if param:match('_text$') or param:match('_valuename_') then --skip commas detection for strings
					pos[param] = value
				elseif param:match('^menu_unlock_') then --store line as is (pure Lua code)
					main.t_unlockLua.modes[param:match('^menu_unlock_(.+)$')] = value
					pos[param] = value
				elseif param:match('^([^_]+)_itemname_') then --skip commas detection and append value to main.t_sort for itemname
					local subt, append = param:match('^([^_]+)_itemname_(.+)$')
					if pos_sort[subt] == nil then
						pos_sort[subt] = {}
					end
					table.insert(pos_sort[subt], append)
					for i = 1, 2 do
						if i == 1 or subt == 'teammenu' then
							local prefix = ''
							if subt == 'teammenu' then
								prefix = 'p' .. i .. '_'
							end
							for _, v in ipairs({'_bg_', '_bg_active_'}) do
								local bg = param:gsub('_itemname_', v)
								def_pos[prefix .. bg .. '_anim'] = -1
								def_pos[prefix .. bg .. '_spr'] = {-1, 0}
								def_pos[prefix .. bg .. '_offset'] = {0, 0}
								def_pos[prefix .. bg .. '_facing'] = 1
								def_pos[prefix .. bg .. '_scale'] = {1.0, 1.0}
							end
						end
					end
					pos[param] = value
				elseif value:match('.+,.+') then --multiple values
					local fontRef = -1
					for i, c in ipairs(main.f_strsplit(',', value)) do --split value using "," delimiter
						if param:match('_anim$') then --mugen recognizes animations even if there are more values
							pos[param] = main.f_dataType(c)
							break
						else
							if i == 1 then
								pos[param] = {}
							end
							if param:match('_font$') then
								-- Change font number reference to font string
								if i == 1 then
									if t.files ~= nil and t.files.font ~= nil and t.files.font[tonumber(c)] ~= nil then
										fontRef = tonumber(c)
										c = t.files.font[fontRef]
									end
								-- Assign default ttf font height declared under [Files], if custom value is not set
								elseif i == 7 and tonumber(c) == -1 and t.files ~= nil and t.files.font_height ~= nil and t.files.font_height[fontRef] ~= nil then
									c = tostring(t.files.font_height[fontRef])
								-- Otherwise validate data
								elseif not tonumber(c) then
									c = nil
								end
							end
						end
						-- Append values
						if c == nil or c == '' then
							table.insert(pos[param], 0)
						else
							table.insert(pos[param], main.f_dataType(c))
						end
					end
				else --single value
					if param:match('_offset$') or param:match('_dist$') or param:match('_speed$') then --precaution in case of optional params without default values
						pos[param] = {}
						table.insert(pos[param], tonumber(value))
					else
						pos[param] = main.f_dataType(value)
					end
				end
			end
		elseif param == nil then --only valid lines left are animations
			line = line:lower()
			local value = line:match('^%s*([0-9%-]+%s*,%s*[0-9%-]+%s*,%s*[0-9%-]+%s*,%s*[0-9%-]+%s*,%s*[0-9%-]+.-)[,%s]*$') or line:match('^%s*loopstart') or line:match('%s*interpolate [oasb][fncl][fgae][sln][ed]t?')
			if value ~= nil then
				value = value:gsub(',%s*,', ',0,') --add missing values
				value = value:gsub(',%s*$', '')
				table.insert(pos, value)
			end
		end
	end
	main.f_loadingRefresh()
end
--file:close()

if main.debugLog then main.f_printTable(main.t_sort, 'debug/t_sort.txt') end

--;===========================================================
--; FIX REFERENCES, LOAD DATA
--;===========================================================
--adopt old DEF code to Ikemen features
if type(t.select_info.cell_spacing) ~= "table" then
	t.select_info.cell_spacing = {t.select_info.cell_spacing, t.select_info.cell_spacing}
end

--training_info section reuses menu_info values (excluding itemnames)
motif.training_info = main.f_tableMerge(motif.training_info, motif.menu_info)
if t.menu_info == nil then t.menu_info = {} end
if t.training_info == nil then t.training_info = {} end
for k, v in pairs(t.menu_info) do
	if t.training_info[k] == nil and not k:match('_itemname_') then
		t.training_info[k] = v
	end
end

--merge tables
motif = main.f_tableMerge(motif, t)

--default hiscore glyphs
if #motif.hiscore_info.glyphs == 0 then
	motif.hiscore_info.glyphs = {'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '!', '?', '.', '<', '>'}
end

--keymenu.item parameters use corresponding menu.item values if not assigned
local t_keymenu = {}
if t.option_info == nil then
	t.option_info = {}
end
for k, v in pairs(motif.option_info) do
	if k:match('^menu_item_') and t.option_info['keymenu' .. k:match('^menu(_item_.+)$')] == nil and motif.option_info['keymenu' .. k:match('^menu(_item_.+)$')] == nil then
		t_keymenu['keymenu' .. k:match('^menu(_item_.+)$')] = v
	end
end
motif.option_info = main.f_tableMerge(motif.option_info, t_keymenu)

if motif.victory_screen.enabled == 0 then
	motif.victory_screen.cpu_enabled = 0
	motif.victory_screen.vs_enabled = 0
end

--adjust window parameters
for k, v in pairs({
	select_info = {'p1_face_window', 'p2_face_window', 'p1_face2_window', 'p2_face2_window', 'stage_portrait_window'},
	vs_screen = {'p1_window', 'p2_window', 'p1_face2_window', 'p2_face2_window'},
	victory_screen = {'p1_window', 'p2_window', 'p1_face2_window', 'p2_face2_window', 'winquote_window'},
	dialogue_info = {'p1_face_window', 'p2_face_window', 'p1_text_window', 'p2_text_window'},
	hiscore_info = {'item_face_window'},
}) do
	for _, param in ipairs(v) do
		--convert mugen style window coordinate system to the one used in engine
		if t[k] == nil or t[k][param] == nil then
			motif[k][param] = {0, 0, motif.info.localcoord[1], motif.info.localcoord[2]}
		else
			motif[k][param][1] = tonumber(motif[k][param][1]) or 0
			motif[k][param][2] = tonumber(motif[k][param][2]) or 0
			motif[k][param][3] = tonumber(motif[k][param][3]) or motif.info.localcoord[1]
			motif[k][param][4] = tonumber(motif[k][param][4]) or motif.info.localcoord[2]
		end
		local window = main.f_tableCopy(motif[k][param])
		if window[3] < window[1] then
			motif[k][param][3] = window[1]
			motif[k][param][1] = window[3]
		end
		if window[4] < window[2] then
			motif[k][param][4] = window[2]
			motif[k][param][2] = window[4]
		end
		if param ~= 'winquote_window' and param ~= 'p1_text_window' and param ~= 'p2_text_window' then
			motif[k][param][3] = motif[k][param][3] - motif[k][param][1]
			motif[k][param][4] = motif[k][param][4] - motif[k][param][2]
		end
	end
end

--general paths
for _, v in ipairs({
	{group = 'files', param = 'spr', dirs = {motif.fileDir, '', 'data/'}},
	{group = 'files', param = 'snd', dirs = {motif.fileDir, '', 'data/'}},
	{group = 'files', param = 'logo_storyboard', dirs = {motif.fileDir, '', 'data/'}},
	{group = 'files', param = 'intro_storyboard', dirs = {motif.fileDir, '', 'data/'}},
	{group = 'files', param = 'select', dirs = {motif.fileDir, '', 'data/'}},
	{group = 'files', param = 'fight', dirs = {motif.fileDir, '', 'data/'}},
	{group = 'files', param = 'glyphs', dirs = {motif.fileDir, '', 'data/'}},
	{group = 'files', param = 'module', dirs = {motif.fileDir, '', 'data/'}},
	{group = 'music', param = 'title_bgm', dirs = {motif.fileDir, '', 'data/', 'sound/'}},
	{group = 'music', param = 'select_bgm', dirs = {motif.fileDir, '', 'data/', 'sound/'}},
	{group = 'music', param = 'vs_bgm', dirs = {motif.fileDir, '', 'data/', 'sound/'}},
	{group = 'music', param = 'victory_bgm', dirs = {motif.fileDir, '', 'data/', 'sound/'}},
	{group = 'music', param = 'option_bgm', dirs = {motif.fileDir, '', 'data/', 'sound/'}},
	{group = 'music', param = 'replay_bgm', dirs = {motif.fileDir, '', 'data/', 'sound/'}},
	{group = 'music', param = 'continue_bgm', dirs = {motif.fileDir, '', 'data/', 'sound/'}},
	{group = 'music', param = 'continue_end_bgm', dirs = {motif.fileDir, '', 'data/', 'sound/'}},
	{group = 'music', param = 'results_bgm', dirs = {motif.fileDir, '', 'data/', 'sound/'}},
	{group = 'music', param = 'hiscore_bgm', dirs = {motif.fileDir, '', 'data/', 'sound/'}},
	{group = 'default_ending', param = 'storyboard', dirs = {motif.fileDir, '', 'data/'}},
	{group = 'end_credits', param = 'storyboard', dirs = {motif.fileDir, '', 'data/'}},
	{group = 'game_over_screen', param = 'storyboard', dirs = {motif.fileDir, '', 'data/'}},
	{group = 'attract_mode', param = 'logo_storyboard', dirs = {motif.fileDir, '', 'data/'}},
	{group = 'attract_mode', param = 'intro_storyboard', dirs = {motif.fileDir, '', 'data/'}},
	{group = 'attract_mode', param = 'start_storyboard', dirs = {motif.fileDir, '', 'data/'}},
}) do
	motif[v.group][v.param] = searchFile(motif[v.group][v.param], v.dirs)
end

motif.files.spr_data = sffNew(motif.files.spr)
main.f_loadingRefresh()
motif.files.snd_data = sndNew(motif.files.snd)
main.f_loadingRefresh()

if main.f_fileExists(motif.files.glyphs) then
	motif.files.glyphs_data = sffNew(motif.files.glyphs)
else
	motif.files.glyphs_data = sffNew()
end
main.f_loadingRefresh()

--motif background data
for k, _ in pairs(motif) do
	if k:match('bgdef$') then
		--optional sff paths and data
		if motif[k].spr ~= nil and motif[k].spr ~= '' then
			motif[k].spr = searchFile(motif[k].spr, {motif.fileDir, '', 'data/'})
			motif[k].spr_data = sffNew(motif[k].spr)
			main.f_loadingRefresh()
		else
			motif[k].spr = motif.files.spr
			motif[k].spr_data = motif.files.spr_data
		end
		--backgrounds
		motif[k].bg = bgNew(motif[k].spr_data, motif.def, k:match('^(.+)def$'))
		main.f_loadingRefresh()
	end
end

--results screens reuse winbgdef values if not defined
for _, v in ipairs{'survivalresultsbgdef', 'timeattackresultsbgdef'} do
	if t[v] == nil then
		motif[v] = motif.winbgdef
	end
end

--trainingbgdef section reuses menubgdef values if not defined
if t.trainingbgdef == nil then
	motif.trainingbgdef = motif.menubgdef
end

--converts facing value to letter used in anim declaration
function motif.f_animFacing(var)
	if var == -1 then
		return 'H'
	else
		return nil
	end
end

--creates sprite data out of table values
local anim = ''
local facing = ''
function motif.f_loadSprData(t, v)
	local animParam = v.s .. 'anim'
	local sprParam = v.s .. 'spr'
	local data = v.s .. 'data'
	-- optional prefix argument only changes parameter name for anim/spr numbers assignment
	if v.prefix ~= nil then
		animParam = v.s .. v.prefix .. 'anim'
		sprParam = v.s .. v.prefix .. 'spr'
		data = v.s .. v.prefix .. 'data'
	end
	if t[v.s .. 'offset'] == nil then t[v.s .. 'offset'] = {0, 0} end
	if t[v.s .. 'scale'] == nil then t[v.s .. 'scale'] = {1.0, 1.0} end
	if t[animParam] ~= nil and t[animParam] ~= -1 and motif.anim[t[animParam]] ~= nil then --create animation data
		if t[v.s .. 'facing'] == nil then t[v.s .. 'facing'] = 1 end
		t[data] = main.f_animFromTable(
			motif.anim[t[animParam]],
			motif.files.spr_data,
			(t[v.s .. 'offset'][1] + (v.x or 0)) / t[v.s .. 'scale'][1],
			(t[v.s .. 'offset'][2] + (v.y or 0)) / t[v.s .. 'scale'][2],
			t[v.s .. 'scale'][1],
			t[v.s .. 'scale'][2],
			motif.f_animFacing(t[v.s .. 'facing'])
		)
	elseif t[sprParam] ~= nil and #t[sprParam] > 0 then --create sprite data
		if #t[sprParam] == 1 then --fix values
			if type(t[sprParam][1]) == 'string' then
				t[sprParam] = {tonumber(t[sprParam][1]:match('^([0-9]+)')), 0}
			else
				t[sprParam] = {t[sprParam][1], 0}
			end
		end
		if t[v.s .. 'facing'] == -1 then facing = ', H' else facing = '' end
		t[data] = animNew(motif.files.spr_data, t[sprParam][1] .. ', ' .. t[sprParam][2] .. ', ' .. (t[v.s .. 'offset'][1] + (v.x or 0)) / t[v.s .. 'scale'][1] .. ', ' .. (t[v.s .. 'offset'][2] + (v.y or 0)) / t[v.s .. 'scale'][2] .. ', -1' .. facing)
		animSetScale(t[data], t[v.s .. 'scale'][1], t[v.s .. 'scale'][2])
		animUpdate(t[data])
	else --create dummy data
		t[data] = animNew(motif.files.spr_data, '-1,0, 0,0, -1')
		animUpdate(t[data])
	end
	animSetWindow(t[data], 0, 0, motif.info.localcoord[1], motif.info.localcoord[2])
end

--creates fadein/fadeout anim data
for k, v in pairs(motif) do
	if type(v) == "table" then
		if motif[k].fadein_anim ~= nil and motif[k].fadein_anim > -1 then
			motif.f_loadSprData(v, {s = 'fadein_'})
		end
		if motif[k].fadeout_anim ~= nil and motif[k].fadeout_anim > -1 then
			motif.f_loadSprData(v, {s = 'fadeout_'})
		end
	end
end

local t_pos = motif.select_info
for _, v in ipairs({
	{s = 'cell_bg_',                      x = 0,                                                           y = 0},
	{s = 'cell_random_',                  x = 0,                                                           y = 0},
	{s = 'p1_teammenu_bg_',               x = t_pos.p1_teammenu_pos[1],                                    y = t_pos.p1_teammenu_pos[2]},
	{s = 'p1_teammenu_selftitle_',        x = t_pos.p1_teammenu_pos[1],                                    y = t_pos.p1_teammenu_pos[2]},
	{s = 'p1_teammenu_enemytitle_',       x = t_pos.p1_teammenu_pos[1],                                    y = t_pos.p1_teammenu_pos[2]},
	{s = 'p1_teammenu_item_cursor_',      x = t_pos.p1_teammenu_pos[1] + t_pos.p1_teammenu_item_offset[1], y = t_pos.p1_teammenu_pos[2] + t_pos.p1_teammenu_item_offset[2]},
	{s = 'p1_teammenu_value_icon_',       x = t_pos.p1_teammenu_pos[1] + t_pos.p1_teammenu_item_offset[1], y = t_pos.p1_teammenu_pos[2] + t_pos.p1_teammenu_item_offset[2]},
	{s = 'p1_teammenu_value_empty_icon_', x = t_pos.p1_teammenu_pos[1] + t_pos.p1_teammenu_item_offset[1], y = t_pos.p1_teammenu_pos[2] + t_pos.p1_teammenu_item_offset[2]},
	{s = 'p1_teammenu_ratio1_icon_',      x = t_pos.p1_teammenu_pos[1] + t_pos.p1_teammenu_item_offset[1], y = t_pos.p1_teammenu_pos[2] + t_pos.p1_teammenu_item_offset[2]},
	{s = 'p1_teammenu_ratio2_icon_',      x = t_pos.p1_teammenu_pos[1] + t_pos.p1_teammenu_item_offset[1], y = t_pos.p1_teammenu_pos[2] + t_pos.p1_teammenu_item_offset[2]},
	{s = 'p1_teammenu_ratio3_icon_',      x = t_pos.p1_teammenu_pos[1] + t_pos.p1_teammenu_item_offset[1], y = t_pos.p1_teammenu_pos[2] + t_pos.p1_teammenu_item_offset[2]},
	{s = 'p1_teammenu_ratio4_icon_',      x = t_pos.p1_teammenu_pos[1] + t_pos.p1_teammenu_item_offset[1], y = t_pos.p1_teammenu_pos[2] + t_pos.p1_teammenu_item_offset[2]},
	{s = 'p1_teammenu_ratio5_icon_',      x = t_pos.p1_teammenu_pos[1] + t_pos.p1_teammenu_item_offset[1], y = t_pos.p1_teammenu_pos[2] + t_pos.p1_teammenu_item_offset[2]},
	{s = 'p1_teammenu_ratio6_icon_',      x = t_pos.p1_teammenu_pos[1] + t_pos.p1_teammenu_item_offset[1], y = t_pos.p1_teammenu_pos[2] + t_pos.p1_teammenu_item_offset[2]},
	{s = 'p1_teammenu_ratio7_icon_',      x = t_pos.p1_teammenu_pos[1] + t_pos.p1_teammenu_item_offset[1], y = t_pos.p1_teammenu_pos[2] + t_pos.p1_teammenu_item_offset[2]},
	{s = 'p2_teammenu_bg_',               x = t_pos.p2_teammenu_pos[1],                                    y = t_pos.p2_teammenu_pos[2]},
	{s = 'p2_teammenu_selftitle_',        x = t_pos.p2_teammenu_pos[1],                                    y = t_pos.p2_teammenu_pos[2]},
	{s = 'p2_teammenu_enemytitle_',       x = t_pos.p2_teammenu_pos[1],                                    y = t_pos.p2_teammenu_pos[2]},
	{s = 'p2_teammenu_item_cursor_',      x = t_pos.p2_teammenu_pos[1] + t_pos.p2_teammenu_item_offset[1], y = t_pos.p2_teammenu_pos[2] + t_pos.p2_teammenu_item_offset[2]},
	{s = 'p2_teammenu_value_icon_',       x = t_pos.p2_teammenu_pos[1] + t_pos.p2_teammenu_item_offset[1], y = t_pos.p2_teammenu_pos[2] + t_pos.p2_teammenu_item_offset[2]},
	{s = 'p2_teammenu_value_empty_icon_', x = t_pos.p2_teammenu_pos[1] + t_pos.p2_teammenu_item_offset[1], y = t_pos.p2_teammenu_pos[2] + t_pos.p2_teammenu_item_offset[2]},
	{s = 'p2_teammenu_ratio1_icon_',      x = t_pos.p2_teammenu_pos[1] + t_pos.p2_teammenu_item_offset[1], y = t_pos.p2_teammenu_pos[2] + t_pos.p2_teammenu_item_offset[2]},
	{s = 'p2_teammenu_ratio2_icon_',      x = t_pos.p2_teammenu_pos[1] + t_pos.p2_teammenu_item_offset[1], y = t_pos.p2_teammenu_pos[2] + t_pos.p2_teammenu_item_offset[2]},
	{s = 'p2_teammenu_ratio3_icon_',      x = t_pos.p2_teammenu_pos[1] + t_pos.p2_teammenu_item_offset[1], y = t_pos.p2_teammenu_pos[2] + t_pos.p2_teammenu_item_offset[2]},
	{s = 'p2_teammenu_ratio4_icon_',      x = t_pos.p2_teammenu_pos[1] + t_pos.p2_teammenu_item_offset[1], y = t_pos.p2_teammenu_pos[2] + t_pos.p2_teammenu_item_offset[2]},
	{s = 'p2_teammenu_ratio5_icon_',      x = t_pos.p2_teammenu_pos[1] + t_pos.p2_teammenu_item_offset[1], y = t_pos.p2_teammenu_pos[2] + t_pos.p2_teammenu_item_offset[2]},
	{s = 'p2_teammenu_ratio6_icon_',      x = t_pos.p2_teammenu_pos[1] + t_pos.p2_teammenu_item_offset[1], y = t_pos.p2_teammenu_pos[2] + t_pos.p2_teammenu_item_offset[2]},
	{s = 'p2_teammenu_ratio7_icon_',      x = t_pos.p2_teammenu_pos[1] + t_pos.p2_teammenu_item_offset[1], y = t_pos.p2_teammenu_pos[2] + t_pos.p2_teammenu_item_offset[2]},
	{s = 'stage_portrait_random_',        x = t_pos.stage_pos[1],                                          y = t_pos.stage_pos[2]},
	{s = 'stage_portrait_bg_',            x = t_pos.stage_pos[1],                                          y = t_pos.stage_pos[2]},
}) do
	motif.f_loadSprData(motif.select_info, v)
end

--versus screen spr/anim data
for i = 1, 2 do
	for j = 1, motif.vs_screen['p' .. i .. '_num'] do
		motif.f_loadSprData(motif.vs_screen, {s = 'p' .. i .. '_member' .. j .. '_icon_'})
		motif.f_loadSprData(motif.vs_screen, {s = 'p' .. i .. '_member' .. j .. '_icon_', prefix = 'done_'})
	end
	for _, v in ipairs({'_value_icon_', '_value_empty_icon_'}) do
		for j = 1, 8 do
			if motif.vs_screen['p' .. i .. v .. 'member' .. j .. '_spr'] ~= nil or motif.vs_screen['p' .. i .. v .. 'member' .. j .. '_anim'] ~= nil then
				motif.f_loadSprData(motif.vs_screen, {s = 'p' .. i .. v, prefix = 'member' .. j .. '_'})
			end
		end
	end
	motif.f_loadSprData(motif.vs_screen, {s = 'p' .. i .. '_value_icon_'})
	motif.f_loadSprData(motif.vs_screen, {s = 'p' .. i .. '_value_empty_icon_'})
end

--continue screen spr/anim data
motif.f_loadSprData(motif.continue_screen, {s = 'counter_'})

--challenger spr/anim data
motif.f_loadSprData(motif.challenger_info, {s = 'bg_'})

--arrows spr/anim data
for _, v in ipairs({motif.title_info, motif.option_info, motif.replay_info, motif.menu_info, motif.training_info, motif.attract_mode}) do
	motif.f_loadSprData(v, {s = 'menu_arrow_up_',   x = v.menu_pos[1], y = v.menu_pos[2]})
	motif.f_loadSprData(v, {s = 'menu_arrow_down_', x = v.menu_pos[1], y = v.menu_pos[2]})
end
for _, v in ipairs({motif.menu_info, motif.training_info}) do
	motif.f_loadSprData(v, {s = 'movelist_arrow_up_',   x = v.movelist_pos[1], y = v.movelist_pos[2]})
	motif.f_loadSprData(v, {s = 'movelist_arrow_down_', x = v.movelist_pos[1], y = v.movelist_pos[2]})
end

--dialogue spr/anim data
for i = 1, 2 do
	motif.f_loadSprData(motif.dialogue_info, {s = 'p' .. i .. '_bg_'})
	motif.f_loadSprData(motif.dialogue_info, {s = 'p' .. i .. '_active_'})
end

--hiscore spr/anim data
motif.f_loadSprData(motif.hiscore_info, {s = 'item_face_bg_'})
motif.f_loadSprData(motif.hiscore_info, {s = 'item_face_unknown_'})

--glyphs spr data
motif.glyphs_data = {}
for k, v in pairs(motif.glyphs) do
	--https://www.ssec.wisc.edu/~tomw/java/unicode.html#xE000
	k = numberToRune(v[1] + 0xe000) --Private Use 0xe000 (57344) - 0xf8ff (63743)
	local anim = animNew(motif.files.glyphs_data, v[1] .. ', ' .. v[2] .. ', 0, 0, -1')
	--animSetScale(anim, 1, 1)
	animUpdate(anim)
	motif.glyphs_data[k] = {
		anim = anim,
		--info = animGetSpriteInfo(anim, v[1], v[2]),
		info = animGetSpriteInfo(anim),
	}
end

-- initialize at the end of main.lua
function motif.f_start()
	-- menus spr/anim data
	for group_k, group_t in pairs(main.t_sort) do
		for subt_k, subt_t in pairs(group_t) do
			for _, v in ipairs(subt_t) do
				if subt_k == 'teammenu' then
					for i = 1, 2 do
						motif.f_loadSprData(motif[group_k], {s = 'p' .. i .. '_' .. subt_k .. '_bg_' .. v .. '_', x = motif[group_k]['p' .. i .. '_teammenu_pos'][1], y = motif[group_k]['p' .. i .. '_teammenu_pos'][2]})
						motif.f_loadSprData(motif[group_k], {s = 'p' .. i .. '_' .. subt_k .. '_bg_active_' .. v .. '_', x = motif[group_k]['p' .. i .. '_teammenu_pos'][1], y = motif[group_k]['p' .. i .. '_teammenu_pos'][2]})
					end
				else--if subt_k == 'menu' or subt_k == 'keymenu' then
					motif.f_loadSprData(motif[group_k], {s = subt_k .. '_bg_' .. v .. '_', x = motif[group_k].menu_pos[1], y = motif[group_k].menu_pos[2]})
					motif.f_loadSprData(motif[group_k], {s = subt_k .. '_bg_active_' .. v .. '_', x = motif[group_k].menu_pos[1], y = motif[group_k].menu_pos[2]})
				end
			end
		end
	end
end

--commands
for _, v in ipairs({
	motif.title_info.menu_next_key,
	motif.title_info.menu_previous_key,
	motif.title_info.menu_accept_key,
	motif.title_info.menu_hiscore_key,
	motif.select_info.p1_teammenu_next_key,
	motif.select_info.p1_teammenu_previous_key,
	motif.select_info.p1_teammenu_add_key,
	motif.select_info.p1_teammenu_subtract_key,
	motif.select_info.p1_teammenu_accept_key,
	motif.select_info.p2_teammenu_next_key,
	motif.select_info.p2_teammenu_previous_key,
	motif.select_info.p2_teammenu_add_key,
	motif.select_info.p2_teammenu_subtract_key,
	motif.select_info.p2_teammenu_accept_key,
	motif.vs_screen.p1_accept_key,
	motif.vs_screen.p1_skip_key,
	motif.vs_screen.p2_accept_key,
	motif.vs_screen.p2_skip_key,
	motif.attract_mode.menu_next_key,
	motif.attract_mode.menu_previous_key,
	motif.attract_mode.menu_accept_key,
	motif.dialogue_info.skip_key,
	motif.dialogue_info.cancel_key,
}) do
	for _, cmd in ipairs (main.f_extractKeys(v)) do
		main.f_commandAdd(cmd, cmd)
	end
end
for i = 1, 2 do
	local j = 1
	while true do
		if motif.vs_screen['p' .. i .. '_member' .. j .. '_key'] == nil then
			break
		end
		for _, cmd in ipairs (main.f_extractKeys(motif.vs_screen['p' .. i .. '_member' .. j .. '_key'])) do
			main.f_commandAdd(cmd, cmd)
		end
		j = j + 1
	end
end

--disabled scaling if element uses default values (non-existing in mugen)
motif.defaultMenu = motif.menu_info.menu_uselocalcoord == 0
motif.defaultOptions = motif.option_info.menu_uselocalcoord == 0
motif.defaultOptionsTitle = t.option_info == nil or t.option_info.title_offset == nil
motif.defaultReplay = motif.replay_info.menu_uselocalcoord == 0
motif.defaultWarning = t.warning_info == nil

if main.debugLog then main.f_printTable(motif, "debug/t_motif.txt") end

return motif
