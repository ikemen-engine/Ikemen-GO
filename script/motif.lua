--;===========================================================
--; DEFAULT VALUES
--;===========================================================
--This pre-made table (3/4 of the whole file) contains all default values used in screenpack. New table from parsed DEF file is merged on top of this one.
--This is important because there are more params available in Ikemen. Whole screenpack code refers to these values.
local motif =
{
	info =
	{
		name = 'Default',
		author = 'Elecbyte',
		versiondate = '09,01,2009',
		mugenversion = '1.0',
		localcoord = {320, 240},
	},
	files =
	{
		spr = 'data/system.sff',
		snd = 'data/system.snd',
		continue_snd = 'data/continue.snd', --Ikemen feature (optional separate entry for better compatibility with existing screenpacks)
		logo_storyboard = '',
		intro_storyboard = '',
		select = 'data/select.def',
		fight = 'data/fight.def',
		debug_font = 'font/f-6x9.fnt', --Ikemen feature
		debug_script = 'script/debug.lua', --Ikemen feature
		font = --FNT v2 fonts not supported yet
		{
			[1] = 'font/f-4x6.fnt',
			[2] = 'font/f-6x9.fnt',
			[3] = 'font/jg.fnt',
		},
		font_height = --Truetype fonts not supported yet
		{
			[1] = nil,
			[2] = nil,
			[3] = nil,
		},
	},
	ja_files = {}, --not used in Ikemen
	music =
	{
		title_bgm = '',
		title_bgm_volume = 100,
		title_bgm_loop = 1, --not supported yet
		title_bgm_loopstart = nil, --not supported yet
		title_bgm_loopend = nil, --not supported yet
		select_bgm = '',
		select_bgm_volume = 100,
		select_bgm_loop = 1, --not supported yet
		select_bgm_loopstart = nil, --not supported yet
		select_bgm_loopend = nil, --not supported yet
		vs_bgm = '',
		vs_bgm_volume = 100,
		vs_bgm_loop = 1, --not supported yet
		vs_bgm_loopstart = nil, --not supported yet
		vs_bgm_loopend = nil, --not supported yet
		victory_bgm = '',
		victory_bgm_volume = 100,
		victory_bgm_loop = 1, --not supported yet
		victory_bgm_loopstart = nil, --not supported yet
		victory_bgm_loopend = nil, --not supported yet
		continue_bgm = 'sound/CONTINUE.ogg', --Ikemen feature
		continue_bgm_volume = 100, --Ikemen feature (not supported yet)
		continue_bgm_loop = 1, --Ikemen feature (not supported yet)
		continue_bgm_loopstart = nil, --Ikemen feature (not supported yet)
		continue_bgm_loopend = nil, --Ikemen feature (not supported yet)
		continue_end_bgm = 'sound/GAME_OVER.ogg', --Ikemen feature
		continue_end_volume = 100, --Ikemen feature (not supported yet)
		continue_end_loop = 0, --Ikemen feature (not supported yet)
		continue_end_loopstart = nil, --Ikemen feature (not supported yet)
		continue_end_loopend = nil, --Ikemen feature (not supported yet)
		results_bgm = '', --Ikemen feature
		results_bgm_loop = 1, --Ikemen feature (not supported yet)
		results_bgm_loopstart = nil, --Ikemen feature (not supported yet)
		results_bgm_loopend = nil, --Ikemen feature (not supported yet)
		tournament_bgm = '', --Ikemen feature
		tournament_bgm_loop = 1, --Ikemen feature (not supported yet)
		tournament_bgm_loopstart = nil, --Ikemen feature (not supported yet)
		tournament_bgm_loopend = nil, --Ikemen feature (not supported yet)
	},
	title_info =
	{
		fadein_time = 10,
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadeout_time = 10,
		fadeout_col = {0, 0, 0}, --Ikemen feature
		loading_offset = {310, 230}, --Ikemen feature
		loading_font = {'font/f-4x6.fnt', 7, -1, nil, nil, nil}, --Ikemen feature
		loading_font_scale = {1.0, 1.0}, --Ikemen feature
		loading_text = 'LOADING...', --Ikemen feature
		footer1_offset = {2, 240}, --Ikemen feature
		footer1_font = {'font/f-4x6.fnt', 7, 1, nil, nil, nil}, --Ikemen feature
		footer1_font_scale = {1.0, 1.0}, --Ikemen feature
		footer1_text = 'I.K.E.M.E.N. by SUEHIRO', --Ikemen feature
		footer2_offset = {160, 240}, --Ikemen feature
		footer2_font = {'font/f-4x6.fnt', 7, 0, nil, nil, nil}, --Ikemen feature
		footer2_font_scale = {1.0, 1.0}, --Ikemen feature
		footer2_text = '', --Ikemen feature
		footer3_offset = {319, 240}, --Ikemen feature
		footer3_font = {'font/f-4x6.fnt', 7, -1, nil, nil, nil}, --Ikemen feature
		footer3_font_scale = {1.0, 1.0}, --Ikemen feature
		footer3_text = 'https://osdn.net/users/supersuehiro/', --Ikemen feature
		footer_boxbackground_visible = 1, --Ikemen feature
		footer_boxbackground_coords = {0, 233, 319, 239}, --Ikemen feature
		footer_boxbackground_alpha = {255, 100}, --Ikemen feature
		connecting_offset = {10, 40}, --Ikemen feature
		connecting_font = {'font/f-6x9.fnt', 0, 1, nil, nil, nil}, --Ikemen feature
		connecting_font_scale = {1.0, 1.0}, --Ikemen feature
		connecting_host = 'Waiting for player 2... (%s)', --Ikemen feature
		connecting_join = 'Now connecting... (%s)', --Ikemen feature
		input_ip_name = 'Enter Host display name, e.g. John.\nExisting entries can be removed with DELETE button.', --Ikemen feature
		input_ip_address = 'Enter Host IP address, e.g. 127.0.0.1\nCopied text can be pasted with INSERT button.', --Ikemen feature
		menu_pos = {159, 158},
		menu_item_font = {'font/f-6x9.fnt', 7, 0, nil, nil, nil},
		menu_item_font_scale = {1.0, 1.0},
		menu_item_active_font = {'font/f-6x9.fnt', 0, 0, nil, nil, nil},
		menu_item_active_font_scale = {1.0, 1.0},
		menu_item_spacing = {0, 13},
		menu_itemname_arcade = 'ARCADE',
		menu_itemname_versus = 'VS MODE',
		menu_itemname_online = 'NETWORK', --Ikemen feature
		menu_itemname_teamarcade = 'TEAM ARCADE', --not used in Ikemen (same as ARCADE)
		menu_itemname_teamversus = 'TEAM VS', --not used in Ikemen (same as VS MODE)
		menu_itemname_teamcoop = 'TEAM CO-OP',
		menu_itemname_survival = 'SURVIVAL',
		menu_itemname_survivalcoop = 'SURVIVAL CO-OP',
		menu_itemname_storymode = 'STORY MODE', --Ikemen feature (not implemented yet)
		menu_itemname_timeattack = 'TIME ATTACK', --Ikemen feature (not implemented yet)
		menu_itemname_tournament = 'TOURNAMENT', --Ikemen feature
		menu_itemname_training = 'TRAINING',
		menu_itemname_watch = 'WATCH',
		menu_itemname_extras = 'EXTRAS', --Ikemen feature
		menu_itemname_options = 'OPTIONS',
		menu_itemname_exit = 'EXIT',
		menu_itemname_netplayversus = 'VS MODE', --Ikemen feature
		menu_itemname_netplayteamcoop = 'TEAM CO-OP', --Ikemen feature
		menu_itemname_netplaysurvivalcoop = 'SURVIVAL CO-OP', --Ikemen feature
		menu_itemname_netplayback = 'BACK', --Ikemen feature
		menu_itemname_freebattle = 'FREE BATTLE', --Ikemen feature
		menu_itemname_timechallenge = 'TIME CHALLENGE', --Ikemen feature (not implemented yet)
		menu_itemname_scorechallenge = 'SCORE CHALLENGE', --Ikemen feature (not implemented yet)
		menu_itemname_100kumite = 'VS 100 KUMITE', --Ikemen feature
		menu_itemname_bossrush = 'BOSS RUSH', --Ikemen feature
		menu_itemname_bonusgames = 'BONUS GAMES', --Ikemen feature
		menu_itemname_scoreranking = 'SCORE RANKING', --Ikemen feature (not implemented yet)
		menu_itemname_replay = 'REPLAY', --Ikemen feature
		menu_itemname_demo = 'DEMO', --Ikemen feature
		menu_itemname_extrasback = 'BACK', --Ikemen feature
		menu_itemname_bonusback = 'BACK', --Ikemen feature
		menu_itemname_tourney32 = 'ROUND OF 32', --Ikemen feature
		menu_itemname_tourney16 = 'ROUND OF 16', --Ikemen feature
		menu_itemname_tourney8 = 'QUARTERFINALS', --Ikemen feature
		menu_itemname_tourney4 = 'SEMIFINALS', --Ikemen feature
		menu_itemname_tourneyback = 'BACK', --Ikemen feature
		menu_itemname_serverhost = 'HOST GAME', --Ikemen feature
		menu_itemname_serverjoin = 'JOIN GAME', --Ikemen feature
		menu_itemname_serverback = 'BACK', --Ikemen feature
		menu_itemname_joinadd = 'NEW ADDRESS', --Ikemen feature
		menu_itemname_joinback = 'BACK', --Ikemen feature
		menu_window_margins_y = {12, 8}, --only partial support for now (menu_window_visibleitems + 1 displayed if y value > 0)
		menu_window_visibleitems = 5,
		menu_boxcursor_visible = 1,
		menu_boxcursor_coords = {-40, -10, 39, 2},
		cursor_move_snd = {100, 0},
		cursor_done_snd = {100, 1},
		cancel_snd = {100, 2},
	},
	titlebgdef =
	{
		spr = '',
		bgclearcolor = {0, 0, 0},
		bg = {},
		timer = 0, --Ikemen feature
	},
	infobox =
	{
		title = 'I.K.E.M.E.N', --Ikemen feature
		title_pos = {159, 13}, --Ikemen feature
		title_font = {'font/f-6x9.fnt', 0, 0, nil, nil, nil}, --Ikemen feature
		title_font_scale = {1.0, 1.0}, --Ikemen feature
		text = "Welcome to I.K.E.M.E.N beta!\n* This is a public development release, for testing purposes.\n* This build isn't stable and may contain bugs and incomplete features.\n* Your help and cooperation are appreciated!\n* Source code: https://osdn.net/users/supersuehiro/\n* Ikemen Plus feedback:\n  http://mugenguild.com/forum/topics/ikemen-plus-181972.200.html", --Ikemen feature (requires new 'text = ' entry under [Infobox] section)
		text_pos = {25, 30}, --Ikemen feature
		text_font = {'font/f-4x6.fnt', 7, 1, nil, nil, nil}, --Ikemen feature
		text_font_scale = {1.0, 1.0}, --Ikemen feature
		text_spacing = {0, 10}, --Ikemen feature
		background_alpha = {20, 100}, --Ikemen feature
	},
	infobox_text = '', --not used in Ikemen
	ja_infobox_text = '', --not used in Ikemen
	select_info =
	{
		fadein_time = 10,
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadeout_time = 10,
		fadeout_col = {0, 0, 0}, --Ikemen feature
		rows = 2,
		columns = 5,
		rows_scrolling = 0, --Ikemen feature
		wrapping = 0,
		wrapping_x = 1, --Ikemen feature
		wrapping_y = 1, --Ikemen feature
		pos = {90, 170},
		double_select = 0, --Ikemen feature
		pos_p1_double_select = {10, 170}, --Ikemen feature
		pos_p2_double_select = {169, 170}, --Ikemen feature
		showemptyboxes = 0,
		moveoveremptyboxes = 0,
		searchemptyboxesup = 0, --Ikemen feature
		searchemptyboxesdown = 0, --Ikemen feature
		cell_size = {27, 27},
		cell_spacing = 2,
		--cell_bg_anim = nil,
		cell_bg_spr = {},
		cell_bg_offset = {0, 0},
		cell_bg_facing = 1,
		cell_bg_scale = {1.0, 1.0},
		--cell_bg_alpha = {}, --Ikemen feature (not implemented yet)
		--cell_random_anim = nil,
		cell_random_spr = {},
		cell_random_offset = {0, 0},
		cell_random_facing = 1,
		cell_random_scale = {1.0, 1.0},
		--cell_random_alpha = {}, --Ikemen feature (not implemented yet)
		cell_random_switchtime = 4,
		p1_cursor_startcell = {0, 0},
		--p1_cursor_active_anim = nil,
		p1_cursor_active_spr = {},
		p1_cursor_active_offset = {0, 0},
		p1_cursor_active_facing = 1,
		p1_cursor_active_scale = {1.0, 1.0},
		--p1_cursor_done_anim = nil,
		p1_cursor_done_spr = {},
		p1_cursor_done_offset = {0, 0},
		p1_cursor_done_facing = 1,
		p1_cursor_done_scale = {1.0, 1.0},
		p1_cursor_move_snd = {100, 0},
		p1_cursor_done_snd = {100, 1},
		p1_random_move_snd = {100, 0},
		p2_cursor_startcell = {0, 4},
		--p2_cursor_active_anim = nil,
		p2_cursor_active_spr = {},
		p2_cursor_active_offset = {0, 0},
		p2_cursor_active_facing = 1,
		p2_cursor_active_scale = {1.0, 1.0},
		--p2_cursor_done_anim = nil,
		p2_cursor_done_spr = {},
		p2_cursor_done_offset = {0, 0},
		p2_cursor_done_facing = 1,
		p2_cursor_done_scale = {1.0, 1.0},
		p2_cursor_blink = 1,
		p2_cursor_move_snd = {100, 0},
		p2_cursor_done_snd = {100, 1},
		p2_random_move_snd = {100, 0},
		random_move_snd_cancel = 0, --not supported yet (needs a function that checks sound length)
		stage_move_snd = {100, 0},
		stage_done_snd = {100, 1},
		cancel_snd = {100, 2},
		portrait_spr = {9000, 0},
		portrait_offset = {0, 0}, --not supported yet
		portrait_scale = {1.0, 1.0},
		title_offset = {0, 0},
		title_font = {'font/jg.fnt', 0, 0, nil, nil, nil},
		title_font_scale = {1.0, 1.0},
		title_text_arcade = 'Arcade', --Ikemen feature
		title_text_versus = 'Versus Mode', --Ikemen feature
		title_text_teamcoop = 'Team Cooperative', --Ikemen feature
		title_text_survival = 'Survival', --Ikemen feature
		title_text_survivalcoop = 'Survival Cooperative', --Ikemen feature
		title_text_storymode = 'Story Mode', --Ikemen feature (not implemented yet)
		title_text_timeattack = 'Time Attack', --Ikemen feature (not implemented yet)
		title_text_training = 'Training Mode', --Ikemen feature
		title_text_watch = 'Watch Mode', --Ikemen feature
		title_text_netplayversus = 'Online Versus', --Ikemen feature
		title_text_netplayteamcoop = 'Online Cooperative', --Ikemen feature
		title_text_netplaysurvivalcoop = 'Online Survival', --Ikemen feature
		title_text_freebattle = 'Free Battle', --Ikemen feature
		title_text_timechallenge = 'Time Challenge', --Ikemen feature (not implemented yet)
		title_text_scorechallenge = 'Score Challenge', --Ikemen feature (not implemented yet)
		title_text_100kumite = 'VS 100 Kumite', --Ikemen feature
		title_text_bossrush = 'Boss Rush', --Ikemen feature
		title_text_replay = 'Replay', --Ikemen feature
		title_text_tourney32 = 'Tournament Mode', --Ikemen feature
		title_text_tourney16 = 'Tournament Mode', --Ikemen feature
		title_text_tourney8 = 'Tournament Mode', --Ikemen feature
		title_text_tourney4 = 'Tournament Mode', --Ikemen feature
		p1_face_spr = {9000, 1},
		p1_face_offset = {0, 0},
		p1_face_facing = 1,
		p1_face_scale = {1.0, 1.0},
		p1_face_window = {0, 0, 0, 0},
		p1_face_spacing = {0, 0}, --Ikemen feature
		p1_face_num = 1, --Ikemen feature
		p2_face_spr = {9000, 1},
		p2_face_offset = {0, 0},
		p2_face_facing = -1,
		p2_face_scale = {1.0, 1.0},
		p2_face_window = {0, 0, 0, 0},
		p2_face_spacing = {0, 0}, --Ikemen feature
		p2_face_num = 1, --Ikemen feature
		p1_name_offset = {0, 0},
		p1_name_font = {'font/jg.fnt', 4, 1, nil, nil, nil},
		p1_name_font_scale = {1.0, 1.0},
		p1_name_spacing = {0, 14},
		p2_name_offset = {0, 0},
		p2_name_font = {'font/jg.fnt', 1, -1, nil, nil, nil},
		p2_name_font_scale = {1.0, 1.0},
		p2_name_spacing = {0, 14},
		stage_pos = {0, 0},
		stage_active_font = {'font/f-4x6.fnt', 0, 0, nil, nil, nil},
		stage_active_font_scale = {1.0, 1.0},
		stage_active2_font = {'font/f-4x6.fnt', 0, 0, nil, nil, nil},
		stage_active2_font_scale = {1.0, 1.0},
		stage_done_font = {'font/f-4x6.fnt', 0, 0, nil, nil, nil},
		stage_done_font_scale = {1.0, 1.0},
		stage_text = 'Stage %i: %s', --Ikemen feature
		stage_text_spacing = {0, 14}, --Ikemen feature
		teammenu_move_wrapping = 1,
		teammenu_itemname_single = 'Single', --Ikemen feature
		teammenu_itemname_simul = 'Simul', --Ikemen feature
		teammenu_itemname_turns = 'Turns', --Ikemen feature
		teammenu_itemname_tag = 'Tag', --Ikemen feature
		p1_teammenu_pos = {0, 0},
		--p1_teammenu_bg_anim = nil,
		p1_teammenu_bg_spr = {},
		p1_teammenu_bg_offset = {0, 0},
		p1_teammenu_bg_facing = 1,
		p1_teammenu_bg_scale = {1.0, 1.0},
		--p1_teammenu_selftitle_anim = nil,
		p1_teammenu_selftitle_spr = {},
		p1_teammenu_selftitle_offset = {0, 0},
		p1_teammenu_selftitle_facing = 1,
		p1_teammenu_selftitle_scale = {1.0, 1.0},
		p1_teammenu_selftitle_font = {'font/jg.fnt', 0, 1, nil, nil, nil},
		p1_teammenu_selftitle_font_scale = {1.0, 1.0},
		p1_teammenu_selftitle_text = '',
		--p1_teammenu_enemytitle_anim = nil,
		p1_teammenu_enemytitle_spr = {},
		p1_teammenu_enemytitle_offset = {0, 0},
		p1_teammenu_enemytitle_facing = 1,
		p1_teammenu_enemytitle_scale = {1.0, 1.0},
		p1_teammenu_enemytitle_font = {'font/jg.fnt', 0, 1, nil, nil, nil},
		p1_teammenu_enemytitle_font_scale = {1.0, 1.0},
		p1_teammenu_enemytitle_text = '',
		p1_teammenu_move_snd = {100, 0},
		p1_teammenu_value_snd = {100, 0},
		p1_teammenu_done_snd = {100, 1},
		p1_teammenu_item_offset = {0, 0},
		p1_teammenu_item_spacing = {0, 15},
		p1_teammenu_item_font_offset = {0, 0}, --Ikemen feature
		p1_teammenu_item_font = {'font/jg.fnt', 0, 1, nil, nil, nil},
		p1_teammenu_item_font_scale = {1.0, 1.0},
		p1_teammenu_item_active_font = {'font/jg.fnt', 3, 1, nil, nil, nil},
		p1_teammenu_item_active_font_scale = {1.0, 1.0},
		p1_teammenu_item_active2_font = {'font/jg.fnt', 0, 1, nil, nil, nil},
		p1_teammenu_item_active2_font_scale = {1.0, 1.0},
		--p1_teammenu_item_cursor_anim = nil,
		p1_teammenu_item_cursor_spr = {},
		p1_teammenu_item_cursor_offset = {0, 0},
		p1_teammenu_item_cursor_facing = 1,
		p1_teammenu_item_cursor_scale = {1.0, 1.0},
		--p1_teammenu_value_icon_anim = nil,
		p1_teammenu_value_icon_spr = {},
		p1_teammenu_value_icon_offset = {0, 0},
		p1_teammenu_value_icon_facing = 1,
		p1_teammenu_value_icon_scale = {1.0, 1.0},
		--p1_teammenu_value_empty_icon_anim = nil,
		p1_teammenu_value_empty_icon_spr = {},
		p1_teammenu_value_empty_icon_offset = {0, 0},
		p1_teammenu_value_empty_icon_facing = 1,
		p1_teammenu_value_empty_icon_scale = {1.0, 1.0},
		p1_teammenu_value_spacing = {6, 0},
		p2_teammenu_pos = {0, 0},
		--p2_teammenu_bg_anim = nil,
		p2_teammenu_bg_spr = {},
		p2_teammenu_bg_offset = {0, 0},
		p2_teammenu_bg_facing = 1,
		p2_teammenu_bg_scale = {1.0, 1.0},
		--p2_teammenu_selftitle_anim = nil,
		p2_teammenu_selftitle_spr = {},
		p2_teammenu_selftitle_offset = {0, 0},
		p2_teammenu_selftitle_facing = 1,
		p2_teammenu_selftitle_scale = {1.0, 1.0},
		p2_teammenu_selftitle_font = {'font/jg.fnt', 0, -1, nil, nil, nil},
		p2_teammenu_selftitle_font_scale = {1.0, 1.0},
		p2_teammenu_selftitle_text = '',
		--p2_teammenu_enemytitle_anim = nil,
		p2_teammenu_enemytitle_spr = {},
		p2_teammenu_enemytitle_offset = {0, 0},
		p2_teammenu_enemytitle_facing = 1,
		p2_teammenu_enemytitle_scale = {1.0, 1.0},
		p2_teammenu_enemytitle_font = {'font/jg.fnt', 0, -1, nil, nil, nil},
		p2_teammenu_enemytitle_font_scale = {1.0, 1.0},
		p2_teammenu_enemytitle_text = '',
		p2_teammenu_move_snd = {100, 0},
		p2_teammenu_value_snd = {100, 0},
		p2_teammenu_done_snd = {100, 1},
		p2_teammenu_item_offset = {0, 0},
		p2_teammenu_item_spacing = {0, 15},
		p2_teammenu_item_font_offset = {0, 0}, --Ikemen feature
		p2_teammenu_item_font = {'font/jg.fnt', 0, -1, nil, nil, nil},
		p2_teammenu_item_font_scale = {1.0, 1.0},
		p2_teammenu_item_active_font = {'font/jg.fnt', 1, -1, nil, nil, nil},
		p2_teammenu_item_active_font_scale = {1.0, 1.0},
		p2_teammenu_item_active2_font = {'font/jg.fnt', 0, -1, nil, nil, nil},
		p2_teammenu_item_active2_font_scale = {1.0, 1.0},
		--p2_teammenu_item_cursor_anim = nil,
		p2_teammenu_item_cursor_spr = {},
		p2_teammenu_item_cursor_offset = {0, 0},
		p2_teammenu_item_cursor_facing = 1,
		p2_teammenu_item_cursor_scale = {1.0, 1.0},
		--p2_teammenu_value_icon_anim = nil,
		p2_teammenu_value_icon_spr = {},
		p2_teammenu_value_icon_offset = {0, 0},
		p2_teammenu_value_icon_facing = 1,
		p2_teammenu_value_icon_scale = {1.0, 1.0},
		--p2_teammenu_value_empty_icon_anim = nil,
		p2_teammenu_value_empty_icon_spr = {},
		p2_teammenu_value_empty_icon_offset = {0, 0},
		p2_teammenu_value_empty_icon_facing = 1,
		p2_teammenu_value_empty_icon_scale = {1.0, 1.0},
		p2_teammenu_value_spacing = {-6, 0},
	},
	selectbgdef =
	{
		spr = '',
		bgclearcolor = {0, 0, 0},
		bg = {},
		timer = 0, --Ikemen feature
	},
	vs_screen =
	{
		fadein_time = 15,
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadeout_time = 15,
		fadeout_col = {0, 0, 0}, --Ikemen feature
		time = 150,
		time_order = 60, --Ikemen feature
		match_text = 'Match %i',
		match_offset = {159, 12},
		match_font = {'font/jg.fnt', 0, 0, nil, nil, nil},
		match_font_scale = {1.0, 1.0},
		p1_pos = {0, 0},
		p1_spr = {9000, 1},
		p1_offset = {0, 0},
		p1_facing = 1,
		p1_scale = {1.0, 1.0},
		p1_window = {0, 0, 0, 0},
		p1_spacing = {0, 0}, --Ikemen feature
		p1_num = 1, --Ikemen feature
		p2_pos = {0, 0},
		p2_spr = {9000, 1},
		p2_offset = {0, 0},
		p2_facing = -1,
		p2_scale = {1.0, 1.0},
		p2_window = {0, 0, 0, 0},
		p2_spacing = {0, 0}, --Ikemen feature
		p2_num = 1, --Ikemen feature
		p1_name_pos = {0, 0},
		p1_name_offset = {0, 0},
		p1_name_font = {'font/jg.fnt', 0, 0, nil, nil, nil},
		p1_name_font_scale = {1.0, 1.0},
		p1_name_spacing = {0, 14},
		p2_name_pos = {0, 0},
		p2_name_offset = {0, 0},
		p2_name_font = {'font/jg.fnt', 0, 0, nil, nil, nil},
		p2_name_font_scale = {1.0, 1.0},
		p2_name_spacing = {0, 14},
		--p1_name_active_font = {'font/jg.fnt', 0, 0, nil, nil, nil}, --Ikemen feature
		--p1_name_active_font_scale = {1.0, 1.0}, --Ikemen feature
		--p2_name_active_font = {'font/jg.fnt', 0, 0, nil, nil, nil}, --Ikemen feature
		--p2_name_active_font_scale = {1.0, 1.0}, --Ikemen feature
		p1_cursor_move_snd = {100, 0}, --Ikemen feature
		p1_cursor_done_snd = {100, 1}, --Ikemen feature
		p2_cursor_move_snd = {100, 0}, --Ikemen feature
		p2_cursor_done_snd = {100, 1}, --Ikemen feature
	},
	versusbgdef =
	{
		spr = '',
		bgclearcolor = {0, 0, 0},
		bg = {},
		timer = 0, --Ikemen feature
	},
	demo_mode =
	{
		enabled = 1, --not supported yet
		select_enabled = 0, --not supported yet
		vsscreen_enabled = 0, --not supported yet
		title_waittime = 600, --not supported yet
		fight_endtime = 1500, --not supported yet
		fight_playbgm = 0, --not supported yet
		fight_stopbgm = 1, --not supported yet
		fight_bars_display = 0, --not supported yet
		intro_waitcycles = 1, --not supported yet
		debuginfo = 0, --not supported yet
	},
	continue_screen =
	{
		external_gameover = 0, --Ikemen feature
		fadein_time = 30, --Ikemen feature
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadeout_time = 30, --Ikemen feature
		fadeout_col = {0, 0, 0}, --Ikemen feature
		endtime = 1593, --Ikemen feature
		continue_starttime = 0, --Ikemen feature
		continue_anim = 9000, --Ikemen feature
		continue_offset = {0, 0}, --Ikemen feature
		continue_scale = {1.0, 1.0}, --Ikemen feature
		continue_skipstart = 71, --Ikemen feature
		continue_9_skiptime = 135, --Ikemen feature
		continue_9_snd = {0, 9}, --Ikemen feature
		continue_8_skiptime = 262, --Ikemen feature
		continue_8_snd = {0, 8}, --Ikemen feature
		continue_7_skiptime = 389, --Ikemen feature
		continue_7_snd = {0, 7}, --Ikemen feature
		continue_6_skiptime = 516, --Ikemen feature
		continue_6_snd = {0, 6}, --Ikemen feature
		continue_5_skiptime = 643, --Ikemen feature
		continue_5_snd = {0, 5}, --Ikemen feature
		continue_4_skiptime = 770, --Ikemen feature
		continue_4_snd = {0, 4}, --Ikemen feature
		continue_3_skiptime = 897, --Ikemen feature
		continue_3_snd = {0, 3}, --Ikemen feature
		continue_2_skiptime = 1024, --Ikemen feature
		continue_2_snd = {0, 2}, --Ikemen feature
		continue_1_skiptime = 1151, --Ikemen feature
		continue_1_snd = {0, 1}, --Ikemen feature
		continue_0_skiptime = 1278, --Ikemen feature
		continue_0_snd = {0, 0}, --Ikemen feature
		continue_end_skiptime = 1366, --Ikemen feature
		continue_end_snd = {1, 0}, --Ikemen feature
		credits_text = 'Credits: %i', --Ikemen feature
		credits_offset = {20, 30}, --Ikemen feature
		credits_font = {'font/jg.fnt', 0, 1, nil, nil, nil}, --Ikemen feature
		credits_font_scale = {1.0, 1.0}, --Ikemen feature
		--enabled = 1, --not used in Ikemen
		--pos = {160, 240}, --not used in Ikemen
		--continue_text = 'CONTINUE?', --not used in Ikemen
		--continue_font = {'font/f-4x6.fnt', 0, 0, nil, nil, nil}, --not used in Ikemen
		--continue_font_scale = {1.0, 1.0}, --not used in Ikemen
		--continue_offset = {0, 0}, --not used in Ikemen
		--yes_text = 'YES', --not used in Ikemen
		--yes_font = {'font/f-4x6.fnt', 0, 0, 128, 128, 128}, --not used in Ikemen
		--yes_font_scale = {1.0, 1.0}, --not used in Ikemen
		--yes_offset = {-80, 60}, --not used in Ikemen
		--yes_active_text = 'YES', --not used in Ikemen
		--yes_active_font = {'font/f-4x6.fnt', 3, 0, nil, nil, nil}, --not used in Ikemen
		--yes_active_font_scale = {1.0, 1.0}, --not used in Ikemen
		--yes_active_offset = {-80, 60}, --not used in Ikemen
		--no_text = 'NO', --not used in Ikemen
		--no_font = {'font/f-4x6.fnt', 0, 0, 128, 128, 128}, --not used in Ikemen
		--no_font_scale = {1.0, 1.0}, --not used in Ikemen
		--no_offset = {80, 60}, --not used in Ikemen
		--no_active_text = 'NO', --not used in Ikemen
		--no_active_font = {'font/f-4x6.fnt', 3, 0, nil, nil, nil}, --not used in Ikemen
		--no_active_font_scale = {1.0, 1.0}, --not used in Ikemen
		--no_active_offset = {80, 60}, --not used in Ikemen
	},
	continuebgdef =
	{
		spr = 'data/continue.sff', --Ikemen feature
		bgclearcolor = {0, 0, 0}, --Ikemen feature
		bg = {},
		timer = 0, --Ikemen feature
	},
	game_over_screen =
	{
		enabled = 1,
		storyboard = '',
	},
	victory_screen =
	{
		enabled = 1,
		time = 300,
		fadein_time = 8,
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadeout_time = 15,
		fadeout_col = {0, 0, 0}, --Ikemen feature
		p1_spr = {9000, 2},
		p1_offset = {100, 20},
		p1_facing = 1,
		p1_scale = {1.0, 1.0},
		p1_window = {0, 0, 319, 160},
		p1_name_offset = {20, 180},
		p1_name_font = {'font/jg.fnt', 0, 1, nil, nil, nil},
		p1_name_font_scale = {1.0, 1.0},
		p2_display = 0, --Ikemen feature
		p2_spr = {9000, 2}, --Ikemen feature
		p2_offset = {100, 20}, --Ikemen feature
		p2_facing = -1, --Ikemen feature
		p2_scale = {1.0, 1.0}, --Ikemen feature
		p2_window = {0, 0, 319, 160}, --Ikemen feature
		p2_name_offset = {20, 180}, --Ikemen feature
		p2_name_font = {'font/jg.fnt', 0, 1, nil, nil, nil}, --Ikemen feature
		p2_name_font_scale = {1.0, 1.0},
		winquote_text = 'Winner!',
		winquote_offset = {20, 192},
		winquote_font = {'font/f-6x9.fnt', 0, 1, nil, nil, nil},
		winquote_font_scale = {1.0, 1.0},
		winquote_spacing = {0, 15}, --Ikemen feature
		winquote_delay = 2, --Ikemen feature
		winquote_length = 50, --Ikemen feature
		winquote_window = {18, 171, 301, 228},
		--winquote_textwrap = 'w', --not used in Ikemen
	},
	victorybgdef =
	{
		spr = '',
		bgclearcolor = {0, 0, 0},
		bg = {},
		timer = 0, --Ikemen feature
	},
	win_screen =
	{
		enabled = 1,
		fadein_time = 32,
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadeout_time = 64,
		fadeout_col = {0, 0, 0}, --Ikemen feature
		pose_time = 300,
		wintext_text = 'Congratulations!',
		wintext_offset = {159, 70},
		wintext_font = {'font/f-6x9.fnt', 0, 0, nil, nil, nil},
		wintext_font_scale = {1.0, 1.0},
		wintext_displaytime = -1,
		wintext_layerno = 2,
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
		fadein_time = 32,
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadeout_time = 64,
		fadeout_col = {0, 0, 0}, --Ikemen feature
		show_time = 300,
		winstext_text = 'Rounds survived: %i',
		winstext_offset = {159, 70},
		winstext_font = {'font/jg.fnt', 0, 0, nil, nil, nil},
		winstext_font_scale = {1.0, 1.0},
		winstext_spacing = {0, 15}, --Ikemen feature
		winstext_displaytime = -1,
		winstext_layerno = 2,
		roundstowin = 5,
	},
	vs100kumite_results_screen =
	{
		enabled = 1, --Ikemen feature
		fadein_time = 32, --Ikemen feature
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadeout_time = 64, --Ikemen feature
		fadeout_col = {0, 0, 0}, --Ikemen feature
		show_time = 300, --Ikemen feature
		winstext_text = 'Wins: %i\nLoses: %i', --Ikemen feature
		winstext_offset = {159, 70}, --Ikemen feature
		winstext_font = {'font/jg.fnt', 0, 0, nil, nil, nil}, --Ikemen feature
		winstext_font_scale = {1.0, 1.0}, --Ikemen feature
		winstext_spacing = {0, 15}, --Ikemen feature
		winstext_displaytime = -1, --Ikemen feature
		winstext_layerno = 2, --Ikemen feature
		roundstowin = 51, --Ikemen feature
	},
	resultsbgdef =
	{
		spr = '', --Ikemen feature
		bgclearcolor = {0, 0, 0}, --Ikemen feature (disabled to not cover game screen)
		bg = {},
		timer = 0, --Ikemen feature
	},
	option_info =
	{
		fadein_time = 10, --check winmugen values
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadeout_time = 10, --check winmugen values
		fadeout_col = {0, 0, 0}, --Ikemen feature
		title_offset = {159, 19},
		title_font = {'font/f-6x9.fnt', 0, 0, nil, nil, nil},
		title_font_scale = {1.0, 1.0},
		title_text_main = 'OPTIONS', --Ikemen feature
		title_text_arcade = 'ARCADE SETTINGS', --Ikemen feature
		title_text_gameplay = 'GAMEPLAY SETTINGS', --Ikemen feature
		title_text_advgameplay = 'ADVANCED GAMEPLAY SETTINGS', --Ikemen feature
		title_text_video = 'VIDEO SETTINGS', --Ikemen feature
		title_text_audio = 'AUDIO SETTINGS', --Ikemen feature
		title_text_engine = 'ENGINE SETTINGS', --Ikemen feature
		title_text_res = 'RESOLUTION SETTINGS', --Ikemen feature
		title_text_input = 'INPUT SETTINGS', --Ikemen feature
		title_text_key = 'KEY SETTINGS', --Ikemen feature
		menu_pos = {85, 33}, --Ikemen feature
		menu_item_font = {'font/f-6x9.fnt', 7, 1, nil, nil, nil}, --Ikemen feature
		menu_item_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_active_font = {'font/f-6x9.fnt', 0, 1, nil, nil, nil}, --Ikemen feature
		menu_item_active_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_value_font = {'font/f-6x9.fnt', 7, -1, nil, nil, nil}, --Ikemen feature
		menu_item_value_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_value_active_font = {'font/f-6x9.fnt', 0, -1, nil, nil, nil}, --Ikemen feature
		menu_item_value_active_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_value_conflict_font = {'font/f-6x9.fnt', 1, -1, nil, nil, nil}, --Ikemen feature
		menu_item_value_conflict_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_spacing = {150, 13}, --Ikemen feature
		menu_window_visibleitems = 16, --Ikemen feature
		menu_boxcursor_visible = 1, --Ikemen feature
		menu_boxcursor_coords = {-5, -9, 154, 3}, --Ikemen feature
		menu_boxbackground_visible = 1, --Ikemen feature
		menu_boxbackground_alpha = {20, 100}, --Ikemen feature
		menu_itemname_main_arcade = 'Arcade Settings', --Ikemen feature
		menu_itemname_main_gameplay = 'Gameplay Settings', --Ikemen feature
		menu_itemname_main_engine = "Engine Settings", --Ikemen feature
		menu_itemname_main_video = 'Video Settings', --Ikemen feature
		menu_itemname_main_audio = 'Audio Settings', --Ikemen feature
		menu_itemname_main_input = 'Input Settings', --Ikemen feature
		menu_itemname_main_port = 'Port Change', --Ikemen feature
		menu_itemname_main_default = 'Default Values', --Ikemen feature
		menu_itemname_main_back = 'Return Without Saving', --Ikemen feature
		menu_itemname_arcade_roundstowin = 'Rounds to Win', --Ikemen feature
		menu_itemname_arcade_roundtime = 'Round Time', --Ikemen feature
		menu_itemname_arcade_difficulty = 'Difficulty level', --Ikemen feature
		menu_itemname_arcade_credits = 'Credits', --Ikemen feature
		menu_itemname_arcade_charchange = 'Char change at Continue', --Ikemen feature
		menu_itemname_arcade_airamping = 'AI ramping', --Ikemen feature
		menu_itemname_arcade_back = 'Back', --Ikemen feature
		menu_itemname_gameplay_lifemul = 'Life', --Ikemen feature
		menu_itemname_gameplay_autoguard = 'Auto-Guard', --Ikemen feature
		menu_itemname_gameplay_attackpowermul = 'Attack.LifeToPowerMul', --Ikemen feature
		menu_itemname_gameplay_gethitpowermul = 'GetHit.LifeToPowerMul', --Ikemen feature
		menu_itemname_gameplay_superdefencemul = 'Super.TargetDefenceMul', --Ikemen feature
		menu_itemname_gameplay_team1vs2life = '1P Vs Team Life', --Ikemen feature
		menu_itemname_gameplay_turnsrecoveryrate = 'Turns HP Recovery', --Ikemen feature
		menu_itemname_gameplay_teampowershare = 'Team Power Share', --Ikemen feature
		menu_itemname_gameplay_teamlifeshare = 'Team Life Share', --Ikemen feature
		menu_itemname_gameplay_singlemode = 'Single Mode', --Ikemen feature
		menu_itemname_gameplay_numturns = 'Turns Limit', --Ikemen feature
		menu_itemname_gameplay_numsimul = 'Simul Limit', --Ikemen feature
		menu_itemname_gameplay_numtag = 'Tag Limit', --Ikemen features
		menu_itemname_gameplay_simulmode = 'Legacy Tag Mode', --Ikemen feature
		menu_itemname_gameplay_simulmode_simul = 'Disabled', --Ikemen feature
		menu_itemname_gameplay_simulmode_tag = 'Enabled', --Ikemen feature
		menu_itemname_gameplay_advanced = 'Advanced Settings', --Ikemen feature
		menu_itemname_gameplay_back = 'Back', --Ikemen feature
		menu_itemname_video_resolution = 'Resolution', --Ikemen feature
		menu_itemname_video_fullscreen = 'Fullscreen', --Ikemen feature
		menu_itemname_video_msaa = 'MSAA', --Ikemen feature
		menu_itemname_video_helpermax = 'HelperMax', --Ikemen feature
		menu_itemname_video_playerprojectilemax = 'PlayerProjectileMax', --Ikemen feature
		menu_itemname_video_explodmax = 'ExplodMax', --Ikemen feature
		menu_itemname_video_afterimagemax = 'AfterImageMax', --Ikemen feature
		menu_itemname_video_zoomactive = 'Zoom Active', --Ikemen feature
		menu_itemname_video_maxzoomout = 'Default Max Zoom Out', --Ikemen feature
		menu_itemname_video_maxzoomin = 'Default Max Zoom In', --Ikemen feature
		menu_itemname_video_zoomspeed = 'Default Zoom Speed', --Ikemen feature
		menu_itemname_video_lifebarfontscale = 'Default Lifebar Font Scale', --Ikemen feature
		menu_itemname_video_aipalette = 'AI Palette', --Ikemen feature
		menu_itemname_video_aipalette_random = 'Random', --Ikemen feature
		menu_itemname_video_aipalette_default = 'Default', --Ikemen feature
		menu_itemname_video_back = 'Back', --Ikemen feature
		menu_itemname_res_320x240 = '320x240    (4:3 QVGA)', --Ikemen feature
		menu_itemname_res_640x480 = '640x480    (4:3 VGA)', --Ikemen feature
		menu_itemname_res_1280x960 = '1280x960   (4:3 Quad-VGA)', --Ikemen feature
		menu_itemname_res_1600x1200 = '1600x1200  (4:3 UXGA)', --Ikemen feature
		menu_itemname_res_960x720 = '960x720    (4:3 HD)', --Ikemen feature
		menu_itemname_res_1200x900 = '1200x900   (4:3 HD+)', --Ikemen feature
		menu_itemname_res_1440x1080 = '1440x1080  (4:3 FHD)', --Ikemen feature
		menu_itemname_res_1280x720 = '1280x720   (16:9 HD)', --Ikemen feature
		menu_itemname_res_1600x900 = '1600x900   (16:9 HD+)', --Ikemen feature
		menu_itemname_res_1920x1080 = '1920x1080  (16:9 FHD)', --Ikemen feature
		menu_itemname_res_2560x1440 = '2560x1440  (16:9 2K)', --Ikemen feature
		menu_itemname_res_3840x2160 = '3840x2160  (16:9 4K)', --Ikemen feature
		menu_itemname_res_custom = 'Custom', --Ikemen feature
		menu_itemname_res_back = 'Back', --Ikemen feature
		menu_itemname_input_p1keyboard = 'P1 Keyboard', --Ikemen feature
		menu_itemname_input_p1gamepad = 'P1 Gamepad', --Ikemen feature
		menu_itemname_input_p2keyboard = 'P2 Keyboard', --Ikemen feature
		menu_itemname_input_p2gamepad = 'P2 Gamepad', --Ikemen feature
		menu_itemname_input_default = 'Default Values', --Ikemen feature
		menu_itemname_input_back = 'Back', --Ikemen feature
		menu_itemname_key_up = 'Up', --Ikemen feature
		menu_itemname_key_down = 'Down', --Ikemen feature
		menu_itemname_key_left = 'Left', --Ikemen feature
		menu_itemname_key_right = 'Right', --Ikemen feature
		menu_itemname_key_a = 'A', --Ikemen feature
		menu_itemname_key_b = 'B', --Ikemen feature
		menu_itemname_key_c = 'C', --Ikemen feature
		menu_itemname_key_x = 'X', --Ikemen feature
		menu_itemname_key_y = 'Y', --Ikemen feature
		menu_itemname_key_z = 'Z', --Ikemen feature
		menu_itemname_key_start = 'Start', --Ikemen feature
		menu_itemname_key_v = 'D', --Ikemen feature
		menu_itemname_key_w = 'W', --Ikemen feature
		menu_itemname_key_back = 'Back', --Ikemen feature
		menu_itemname_yes = 'Yes', --Ikemen feature
		menu_itemname_no = 'No', --Ikemen feature
		menu_itemname_main_save = 'Save and Return', --Ikemen feature
		menu_itemname_enabled = 'Enabled', --Ikemen feature
		menu_itemname_disabled = 'Disabled', --Ikemen feature
		menu_itemname_audio_mastervolume = 'Master Volume', --Ikemen feature
		menu_itemname_audio_sfxvolume = 'SFX Volume', --Ikemen feature
		menu_itemname_audio_bgmvolume = 'BGM Volume', --Ikemen feature
		menu_itemname_audio_audioducking = 'Audio Ducking', --Ikemen feature
		menu_itemname_engine_quicklaunch = 'Quick Launch', --Ikemen feature
		menu_itemname_engine_allowdebugkeys = 'Debug Keys', --Ikemen feature
		input_text_port = 'Type in Host Port, e.g. 7500.\nPress ENTER to accept.\nPress ESC to cancel.', --Ikemen feature
		input_text_reswidth = 'Type in screen width.\nPress ENTER to accept.\nPress ESC to cancel.', --Ikemen feature
		input_text_resheight = 'Type in screen height.\nPress ENTER to accept.\nPress ESC to cancel.', --Ikemen feature
		input_text_key = 'Press a key to assign to entry.\nPress ESC to cancel.', --Ikemen feature
		cursor_move_snd = {100, 0},
		cursor_done_snd = {100, 1},
		cancel_snd = {100, 2},
	},
	optionbgdef =
	{
		spr = '',
		bgclearcolor = {0, 0, 0},
		bg = {},
		timer = 0, --Ikemen feature
	},
	tournament_info =
	{
		fadein_time = 15, --Ikemen feature
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadeout_time = 15, --Ikemen feature
		fadeout_col = {0, 0, 0}, --Ikemen feature
	},
	tournamentbgdef =
	{
		spr = 'data/tournament.sff', --Ikemen feature
		bgclearcolor = {0, 0, 0}, --Ikemen feature
		bg = {}, --Ikemen feature
		timer = 0, --Ikemen feature
	},
	warning_info =
	{
		title = 'WARNING', --Ikemen feature
		title_pos = {159, 19}, --Ikemen feature
		title_font = {'font/f-6x9.fnt', 0, 0, nil, nil, nil}, --Ikemen feature
		title_font_scale = {1.0, 1.0}, --Ikemen feature
		text_stages = 'No stages in select.def available for random selection.\nPress any key to exit the program.', --Ikemen feature
		text_order = "No characters in select.def correspond to 'maxmatches'\nsettings. Check [Characters] section and 'order' parameters.\nPress any key to exit the program.", --Ikemen feature
		text_training = "Training character ('chars/Training/Training.def') not found.\nPress any key to exit the program.", --Ikemen feature
		text_reload = 'Some selected options require Ikemen to be restarted.\nPress any key to exit the program.', --Ikemen feature
		text_res = 'Non 4:3 resolutions require stages coded for different\naspect ratio. Change it back to 4:3 if stages look off.', --Ikemen feature
		text_pos = {25, 33}, --Ikemen feature
		text_font = {'font/f-6x9.fnt', 0, 1, nil, nil, nil}, --Ikemen feature
		text_font_scale = {1.0, 1.0}, --Ikemen feature
		text_spacing = {0, 13}, --Ikemen feature
		background_alpha = {20, 100}, --Ikemen feature
	},
	anim =
	{
		[9000] = { --Ikemen feature
			'2,0, -60,8, 1', --1 (count 9 start - initial animation)
			'2,1, -60,8, 1', --2
			'2,2, -60,8, 1', --3
			'2,3, -60,8, 3', --6
			'2,4, -60,8, 3', --9
			'2,5, -60,8, 1', --10
			'2,6, -60,8, 3', --13
			'2,7, -60,8, 2', --15
			'2,8, -60,8, 2', --17
			'2,9, -60,8, 2', --19
			'2,10, -60,8, 1', --20
			'2,11, -60,8, 5', --25
			'2,12, -60,8, 5', --30
			'2,13, -60,8, 1', --31
			'2,14, -60,8, 1', --32
			'2,15, -60,8, 3', --35
			'2,16, -60,8, 3', --38
			'2,17, -60,8, 1', --39
			'2,18, -60,8, 5', --44
			'2,19, -60,8, 1', --45
			'2,20, -60,8, 4', --49
			'2,21, -60,8, 1', --50
			'2,22, -60,8, 1', --51
			'2,23, -60,8, 1', --52
			'2,24, -60,8, 2', --54
			'2,25, -60,8, 1', --55
			'2,26, -60,8, 5', --60
			'2,27, -60,8, 5', --65
			'2,28, -60,8, 1', --66
			'2,29, -60,8, 3', --69
			'2,30, -60,8, 2', --71
			'3,0, -60,8, 6', --77 (count 9 - actual counter)
			'3,1, -60,8, 54', --131
			'3,2, -60,8, 4', --135
			'3,3, -60,8, 4', --139
			'3,4, -60,8, 4', --143
			'3,5, -60,8, 4', --147
			'3,6, -60,8, 44', --191
			'3,7, -60,8, 9', --200
			'3,8, -60,8, 4', --204
			'3,9, -60,8, 3', --207
			'3,10, -60,8, 2', --209
			'3,11, -60,8, 2', --211
			'3,12, -60,8, 1', --212
			'3,13, -60,8, 2', --214
			'3,14, -60,8, 1', --215
			'3,15, -60,8, 1', --216
			'3,16, -60,8, 2', --218
			'3,17, -60,8, 1', --219
			'3,18, -60,8, 1', --220
			'3,19, -60,8, 1', --221 (count 9 end)
			'4,0, -60,8, 1', --222 (count 8 start)
			'4,1, -60,8, 2', --224
			'4,2, -60,8, 2', --226
			'4,3, -60,8, 2', --228
			'4,4, -60,8, 2', --230
			'4,5, -60,8, 2', --232
			'4,6, -60,8, 2', --234
			'4,7, -60,8, 1', --235
			'4,8, -60,8, 2', --237
			'4,9, -60,8, 3', --240
			'4,10, -60,8, 4', --244
			'4,11, -60,8, 9', --253
			'4,12, -60,8, 5', --258
			'4,13, -60,8, 4', --262
			'4,14, -60,8, 4', --266
			'4,15, -60,8, 4', --270
			'4,16, -60,8, 4', --274
			'4,17, -60,8, 44', --318
			'4,18, -60,8, 9', --327
			'4,19, -60,8, 4', --331
			'4,20, -60,8, 3', --334
			'4,21, -60,8, 2', --336
			'4,22, -60,8, 2', --338
			'4,23, -60,8, 1', --339
			'4,24, -60,8, 2', --341
			'4,25, -60,8, 1', --342
			'4,26, -60,8, 1', --343
			'4,27, -60,8, 2', --345
			'4,28, -60,8, 1', --346
			'4,29, -60,8, 1', --347
			'4,30, -60,8, 1', --348 (count 8 end)
			'5,0, -60,8, 1', --349 (count 7 start)
			'5,1, -60,8, 2', --351
			'5,2, -60,8, 2', --353
			'5,3, -60,8, 2', --355
			'5,4, -60,8, 2', --357
			'5,5, -60,8, 2', --359
			'5,6, -60,8, 2', --361
			'5,7, -60,8, 1', --362
			'5,8, -60,8, 2', --364
			'5,9, -60,8, 3', --367
			'5,10, -60,8, 4', --371
			'5,11, -60,8, 9', --380
			'5,12, -60,8, 5', --385
			'5,13, -60,8, 4', --389
			'5,14, -60,8, 4', --393
			'5,15, -60,8, 4', --397
			'5,16, -60,8, 4', --401
			'5,17, -60,8, 44', --445
			'5,18, -60,8, 9', --454
			'5,19, -60,8, 4', --458
			'5,20, -60,8, 3', --461
			'5,21, -60,8, 2', --463
			'5,22, -60,8, 2', --465
			'5,23, -60,8, 1', --466
			'5,24, -60,8, 2', --468
			'5,25, -60,8, 1', --469
			'5,26, -60,8, 1', --470
			'5,27, -60,8, 2', --472
			'5,28, -60,8, 1', --473
			'5,29, -60,8, 1', --474
			'5,30, -60,8, 1', --475 (count 7 end)
			'6,0, -60,8, 1', --476 (count 6 start)
			'6,1, -60,8, 2', --478
			'6,2, -60,8, 2', --480
			'6,3, -60,8, 2', --482
			'6,4, -60,8, 2', --484
			'6,5, -60,8, 2', --486
			'6,6, -60,8, 2', --488
			'6,7, -60,8, 1', --489
			'6,8, -60,8, 2', --491
			'6,9, -60,8, 3', --494
			'6,10, -60,8, 4', --498
			'6,11, -60,8, 9', --507
			'6,12, -60,8, 5', --512
			'6,13, -60,8, 4', --516
			'6,14, -60,8, 4', --520
			'6,15, -60,8, 4', --524
			'6,16, -60,8, 4', --528
			'6,17, -60,8, 44', --572
			'6,18, -60,8, 9', --581
			'6,19, -60,8, 4', --585
			'6,20, -60,8, 3', --588
			'6,21, -60,8, 2', --590
			'6,22, -60,8, 2', --592
			'6,23, -60,8, 1', --593
			'6,24, -60,8, 2', --595
			'6,25, -60,8, 1', --596
			'6,26, -60,8, 1', --597
			'6,27, -60,8, 2', --599
			'6,28, -60,8, 1', --600
			'6,29, -60,8, 1', --601
			'6,30, -60,8, 1', --602 (count 6 end)
			'7,0, -60,8, 1', --603 (count 5 start)
			'7,1, -60,8, 2', --605
			'7,2, -60,8, 2', --607
			'7,3, -60,8, 2', --609
			'7,4, -60,8, 2', --611
			'7,5, -60,8, 2', --613
			'7,6, -60,8, 2', --615
			'7,7, -60,8, 1', --616
			'7,8, -60,8, 2', --618
			'7,9, -60,8, 3', --621
			'7,10, -60,8, 4', --625
			'7,11, -60,8, 9', --634
			'7,12, -60,8, 5', --639
			'7,13, -60,8, 4', --643
			'7,14, -60,8, 4', --647
			'7,15, -60,8, 4', --651
			'7,16, -60,8, 4', --655
			'7,17, -60,8, 44', --699
			'7,18, -60,8, 9', --708
			'7,19, -60,8, 4', --712
			'7,20, -60,8, 3', --715
			'7,21, -60,8, 2', --717
			'7,22, -60,8, 2', --719
			'7,23, -60,8, 1', --720
			'7,24, -60,8, 2', --722
			'7,25, -60,8, 1', --723
			'7,26, -60,8, 1', --724
			'7,27, -60,8, 2', --726
			'7,28, -60,8, 1', --727
			'7,29, -60,8, 1', --728
			'7,30, -60,8, 1', --729 (count 5 end)
			'8,0, -60,8, 1', --730 (count 4 start)
			'8,1, -60,8, 2', --732
			'8,2, -60,8, 2', --734
			'8,3, -60,8, 2', --736
			'8,4, -60,8, 2', --738
			'8,5, -60,8, 2', --740
			'8,6, -60,8, 2', --742
			'8,7, -60,8, 1', --743
			'8,8, -60,8, 2', --745
			'8,9, -60,8, 3', --748
			'8,10, -60,8, 4', --752
			'8,11, -60,8, 9', --761
			'8,12, -60,8, 5', --766
			'8,13, -60,8, 4', --770
			'8,14, -60,8, 4', --774
			'8,15, -60,8, 4', --778
			'8,16, -60,8, 4', --782
			'8,17, -60,8, 44', --826
			'8,18, -60,8, 9', --835
			'8,19, -60,8, 4', --839
			'8,20, -60,8, 3', --842
			'8,21, -60,8, 2', --844
			'8,22, -60,8, 2', --846
			'8,23, -60,8, 1', --847
			'8,24, -60,8, 2', --849
			'8,25, -60,8, 1', --850
			'8,26, -60,8, 1', --851
			'8,27, -60,8, 2', --853
			'8,28, -60,8, 1', --854
			'8,29, -60,8, 1', --855
			'8,30, -60,8, 1', --856 (count 4 end)
			'9,0, -60,8, 1', --857 (count 3 start)
			'9,1, -60,8, 2', --859
			'9,2, -60,8, 2', --861
			'9,3, -60,8, 2', --863
			'9,4, -60,8, 2', --865
			'9,5, -60,8, 2', --867
			'9,6, -60,8, 2', --869
			'9,7, -60,8, 1', --870
			'9,8, -60,8, 2', --872
			'9,9, -60,8, 3', --875
			'9,10, -60,8, 4', --879
			'9,11, -60,8, 9', --888
			'9,12, -60,8, 5', --893
			'9,13, -60,8, 4', --897
			'9,14, -60,8, 4', --901
			'9,15, -60,8, 4', --905
			'9,16, -60,8, 4', --909
			'9,17, -60,8, 44', --953
			'9,18, -60,8, 9', --962
			'9,19, -60,8, 4', --966
			'9,20, -60,8, 3', --969
			'9,21, -60,8, 2', --971
			'9,22, -60,8, 2', --973
			'9,23, -60,8, 1', --974
			'9,24, -60,8, 2', --976
			'9,25, -60,8, 1', --977
			'9,26, -60,8, 1', --978
			'9,27, -60,8, 2', --980
			'9,28, -60,8, 1', --981
			'9,29, -60,8, 1', --982
			'9,30, -60,8, 1', --983 (count 3 end)
			'1-60,8, -60,8, 1', --984 (count 2 start)
			'10,1, -60,8, 2', --986
			'10,2, -60,8, 2', --988
			'10,3, -60,8, 2', --990
			'10,4, -60,8, 2', --992
			'10,5, -60,8, 2', --994
			'10,6, -60,8, 2', --996
			'10,7, -60,8, 1', --997
			'10,8, -60,8, 2', --999
			'10,9, -60,8, 3', --1002
			'10,10, -60,8, 4', --1006
			'10,11, -60,8, 9', --1015
			'10,12, -60,8, 5', --1020
			'10,13, -60,8, 4', --1024
			'10,14, -60,8, 4', --1028
			'10,15, -60,8, 4', --1032
			'10,16, -60,8, 4', --1036
			'10,17, -60,8, 44', --1080
			'10,18, -60,8, 9', --1089
			'10,19, -60,8, 4', --1093
			'10,20, -60,8, 3', --1096
			'10,21, -60,8, 2', --1098
			'10,22, -60,8, 2', --1100
			'10,23, -60,8, 1', --1101
			'10,24, -60,8, 2', --1103
			'10,25, -60,8, 1', --1104
			'10,26, -60,8, 1', --1105
			'10,27, -60,8, 2', --1107
			'10,28, -60,8, 1', --1108
			'10,29, -60,8, 1', --1109
			'10,30, -60,8, 1', --1110 (count 2 end)
			'11,0, -60,8, 1', --1111 (count 1 start)
			'11,1, -60,8, 2', --1113
			'11,2, -60,8, 2', --1115
			'11,3, -60,8, 2', --1117
			'11,4, -60,8, 2', --1119
			'11,5, -60,8, 2', --1121
			'11,6, -60,8, 2', --1123
			'11,7, -60,8, 1', --1124
			'11,8, -60,8, 2', --1126
			'11,9, -60,8, 3', --1129
			'11,10, -60,8, 4', --1133
			'11,11, -60,8, 9', --1142
			'11,12, -60,8, 5', --1147
			'11,13, -60,8, 4', --1151
			'11,14, -60,8, 4', --1155
			'11,15, -60,8, 4', --1159
			'11,16, -60,8, 4', --1163
			'11,17, -60,8, 44', --1207
			'11,18, -60,8, 9', --1216
			'11,19, -60,8, 4', --1220
			'11,20, -60,8, 3', --1223
			'11,21, -60,8, 2', --1225
			'11,22, -60,8, 2', --1227
			'11,23, -60,8, 1', --1228
			'11,24, -60,8, 2', --1230
			'11,25, -60,8, 1', --1231
			'11,26, -60,8, 1', --1232
			'11,27, -60,8, 2', --1234
			'11,28, -60,8, 1', --1235
			'11,29, -60,8, 1', --1236
			'11,30, -60,8, 1', --1237 (count 1 end)
			'12,0, -60,8, 1', --1238 (count 0 start)
			'12,1, -60,8, 2', --1240
			'12,2, -60,8, 2', --1242
			'12,3, -60,8, 2', --1244
			'12,4, -60,8, 2', --1246
			'12,5, -60,8, 2', --1248
			'12,6, -60,8, 2', --1250
			'12,7, -60,8, 1', --1251
			'12,8, -60,8, 2', --1253
			'12,9, -60,8, 3', --1256
			'12,10, -60,8, 4', --1260
			'12,11, -60,8, 9', --1269
			'12,12, -60,8, 5', --1274
			'12,13, -60,8, 4', --1278
			'12,14, -60,8, 4', --1282
			'12,15, -60,8, 4', --1286
			'12,16, -60,8, 4', --1290
			'12,17, -60,8, 75', --1365
			'12,18, -60,8, 1', --1366  (count 0 end)
			'10-32,8, -32,8, 1', --1367 (1) (game over start)
			'100,1, -32,8, 1', --1368 (2)
			'100,2, -32,8, 1', --1369 (3)
			'100,3, -32,8, 1', --1370 (4)
			'100,4, -32,8, 1', --1371 (5)
			'100,5, -32,8, 1', --1372 (6)
			'100,6, -32,8, 1', --1373 (7)
			'100,7, -32,8, 1', --1374 (8)
			'100,8, -32,8, 1', --1375 (9)
			'100,9, -32,8, 1', --1376 (10)
			'100,10, -32,8, 1', --1377 (11)
			'100,11, -32,8, 1', --1378 (12)
			'100,12, -32,8, 1', --1379 (13)
			'100,13, -32,8, 1', --1380 (14)
			'100,14, -32,8, 1', --1381 (15)
			'100,15, -32,8, 1', --1382 (16)
			'100,16, -32,8, 1', --1383 (17)
			'100,17, -32,8, 1', --1384 (18)
			'100,18, -32,8, 1', --1385 (19)
			'100,19, -32,8, 1', --1386 (20)
			'100,20, -32,8, 1', --1387 (21)
			'100,21, -32,8, 1', --1388 (22)
			'100,22, -32,8, 1', --1389 (23)
			'100,23, -32,8, 1', --1390 (24)
			'100,24, -32,8, 1', --1391 (25)
			'100,25, -32,8, 1', --1392 (26)
			'100,26, -32,8, 1', --1393 (27)
			'100,27, -32,8, 1', --1394 (28)
			'100,28, -32,8, 1', --1395 (29)
			'100,29, -32,8, 1', --1396 (30)
			'100,30, -32,8, 1', --1397 (31)
			'100,31, -32,8, 1', --1398 (32)
			'101,0, -32,8, 1', --1399 (33)
			'101,1, -32,8, 1', --1400 (34)
			'101,2, -32,8, 1', --1401 (35)
			'101,3, -32,8, 1', --1402 (36)
			'101,4, -32,8, 1', --1403 (37)
			'101,5, -32,8, 1', --1404 (38)
			'101,6, -32,8, 1', --1405 (39)
			'101,7, -32,8, 1', --1406 (40)
			'101,8, -32,8, 1', --1407 (41)
			'101,9, -32,8, 1', --1408 (42)
			'101,10, -32,8, 1', --1409 (43)
			'101,11, -32,8, 1', --1410 (44)
			'101,12, -32,8, 1', --1411 (45)
			'101,13, -32,8, 1', --1412 (46)
			'101,14, -32,8, 1', --1413 (47)
			'101,15, -32,8, 1', --1414 (48)
			'101,16, -32,8, 1', --1415 (49)
			'101,17, -32,8, 1', --1416 (50)
			'101,18, -32,8, 1', --1417 (51)
			'101,19, -32,8, 1', --1418 (52)
			'101,20, -32,8, 1', --1419 (53)
			'101,21, -32,8, 1', --1420 (54)
			'101,22, -32,8, 1', --1421 (55)
			'101,23, -32,8, 1', --1422 (56)
			'101,24, -32,8, 1', --1423 (57)
			'101,25, -32,8, 1', --1424 (58)
			'101,26, -32,8, 1', --1425 (59)
			'101,27, -32,8, 1', --1426 (60)
			'101,28, -32,8, 1', --1427 (61)
			'101,29, -32,8, 1', --1428 (62)
			'101,30, -32,8, 1', --1429 (63)
			'101,31, -32,8, 1', --1430 (64)
			'101,32, -32,8, 1', --1431 (65)
			'101,33, -32,8, 1', --1432 (66)
			'101,34, -32,8, 1', --1433 (67)
			'101,35, -32,8, 1', --1434 (68)
			'101,36, -32,8, 1', --1435 (69)
			'101,37, -32,8, 1', --1436 (70)
			'101,38, -32,8, 1', --1437 (71)
			'101,39, -32,8, 1', --1438 (72)
			'101,40, -32,8, 1', --1439 (73)
			'101,41, -32,8, 1', --1440 (74)
			'101,42, -32,8, 1', --1441 (75)
			'101,43, -32,8, 1', --1442 (76)
			'101,44, -32,8, 1', --1443 (77)
			'101,45, -32,8, 1', --1444 (78)
			'101,46, -32,8, 1', --1445 (79)
			'101,47, -32,8, 1', --1446 (80)
			'101,48, -32,8, 1', --1447 (81)
			'101,49, -32,8, 1', --1448 (82)
			'101,50, -32,8, 1', --1449 (83)
			'101,51, -32,8, 1', --1450 (84)
			'101,52, -32,8, 1', --1451 (85)
			'101,53, -32,8, 1', --1452 (86)
			'101,54, -32,8, 1', --1453 (87)
			'101,55, -32,8, 1', --1454 (88)
			'101,56, -32,8, 1', --1455 (89)
			'101,57, -32,8, 1', --1456 (90)
			'101,58, -32,8, 1', --1457 (91)
			'101,59, -32,8, 1', --1458 (92)
			'101,60, -32,8, 1', --1459 (93)
			'101,61, -32,8, 1', --1460 (94)
			'101,62, -32,8, 1', --1461 (95)
			'101,63, -32,8, 1', --1462 (96)
			'101,64, -32,8, 1', --1463 (97)
			'101,65, -32,8, 1', --1464 (98)
			'101,66, -32,8, 1', --1465 (99)
			'101,67, -32,8, 1', --1466 (100)
			'101,68, -32,8, 1', --1467 (101)
			'101,69, -32,8, 1', --1468 (102)
			'101,70, -32,8, 1', --1469 (103)
			'101,71, -32,8, 1', --1470 (104)
			'101,72, -32,8, 1', --1471 (105)
			'101,73, -32,8, 1', --1472 (106)
			'101,74, -32,8, 1', --1473 (107)
			'101,75, -32,8, 119', --1592 (226)
			'101,75, -32,8, -1', --1593+
		},
	},
	ctrldef = {
		titlebgdef = {},
		selectbgdef = {}, 
		versusbgdef = {},
		optionbgdef = {},
		continuebgdef = {},
		victorybgdef = {},
		resultsbgdef = {},
		tournamentbgdef = {},
	}
}

--;===========================================================
--; PARSE SCREENPACK
--;===========================================================
--here starts proper screenpack DEF file parsing
local sp = config.Motif
if main.flags['-r'] ~= nil then
	if main.f_fileExists(main.flags['-r']) then
		sp = main.flags['-r']
	elseif main.f_fileExists('data/' .. main.flags['-r'] .. '/system.def') then
		sp = 'data/' .. main.flags['-r'] .. '/system.def'
	end
end
local file = io.open(sp, 'r')
local fileDir, fileName = sp:match('^(.-)([^/\\]+)$')
local t = {}
local pos = t
local def_pos = motif
t.anim = {}
t.font_data = {['font/f-4x6.fnt'] = fontNew('font/f-4x6.fnt'), ['font/f-6x9.fnt'] = fontNew('font/f-6x9.fnt'), ['font/jg.fnt'] = fontNew('font/jg.fnt')}
t.ctrldef = {}
t.fileDir = fileDir
t.fileName = fileName
local bgdef = 'dummyUntilSet'
local bgctrl = ''
local bgctrl_match = 'dummyUntilSet'
local tmp = ''
for line in file:lines() do
	line = line:gsub('%s*;.*$', '')
	if line:match('^%s*%[.-%s*%]%s*$') then --matched [] group
		line = line:match('^%s*%[(.-)%s*%]%s*$') --match text between []
		line = line:gsub('[%. ]', '_') --change . and space to _
		line = line:lower() --lowercase line
		local row = tostring(line:lower()) --just in case it's a number (not really needed)
		if row:match('.+ctrldef') then --matched ctrldef start
			bgctrl = row
			bgctrl_match = bgctrl:match('^(.-ctrl)def')
			t.ctrldef[bgdef .. 'def'][bgctrl] = {}
			t.ctrldef[bgdef .. 'def'][bgctrl].ctrl = {}
			pos = t.ctrldef[bgdef .. 'def'][bgctrl]
			motif.ctrldef[bgdef .. 'def'][bgctrl] = {
				looptime = -1,
				ctrlid = {0},
				ctrl = {}
			}
		elseif row:match('^' .. bgctrl_match) then --matched ctrldef content
			tmp = t.ctrldef[bgdef .. 'def'][bgctrl].ctrl
			tmp[#tmp + 1] = {}
			pos = tmp[#tmp]
			motif.ctrldef[bgdef .. 'def'][bgctrl].ctrl[#tmp] = {
				type = 'null',
				time = {0, -1, -1},
				ctrlid = {}
			}
		elseif row:match('.+bgdef$') then --matched bgdef start
			t[row] = {}
			pos = t[row]
			t[row].bg = {}
			bgdef = row:match('(.+)def$')
			t.ctrldef[bgdef .. 'def'] = {}
		elseif row:match('^' .. bgdef) then --matched bgdef content
			tmp = t[bgdef .. 'def']
			tmp.bg[#tmp.bg + 1] = {}
			pos = tmp.bg[#tmp.bg]
			motif[bgdef .. 'def'].bg[#tmp.bg] =
			{
				type = 'normal',
				spriteno = {0, 0},
				id = 0,
				layerno = 0,
				start = {0, 0},
				delta = {1, 1},
				trans = '',
				mask = 0,
				tile = {0, 0},
				tilespacing = {0, nil},
				--window = {0, 0, 0, 0},
				--windowdelta = {0, 0}, --not supported yet
				--width = {0, 0}, --not supported yet (parallax)
				--xscale = {1.0, 1.0}, --not supported yet (parallax)
				--yscalestart = 100, --not supported yet (parallax)
				--yscaledelta = 1, --not supported yet (parallax)
				positionlink = 0,
				velocity = {0, 0},
				--sin_x = {0, 0, 0}, --not supported yet
				--sin_y = {0, 0, 0}, --not supported yet
				ctrl = {},
				ctrl_flags = {
					visible = 1,
					enabled = 1,
					velx = 0,
					vely = 0,
					x = 0,
					y = 0
				}
			}
		elseif row:match('^begin_action_[0-9]+$') then --matched anim
			row = tonumber(row:match('^begin_action_([0-9]+)$'))
			t.anim[row] = {}
			pos = t.anim[row]
		else --matched other []
			t[row] = {}
			pos = t[row]
			def_pos = motif[row]
		end
	else --matched non [] line
		local param, value = line:match('^%s*([^=]-)%s*=%s*(.-)%s*$')
		if param ~= nil then
			param = param:gsub('[%. ]', '_') --change param . and space to _
			param = param:lower() --lowercase param
			if value == '' and (type(def_pos[param]) == 'number' or type(def_pos[param]) == 'table') then --text should remain empty
				value = '0'
			end
		end
		if param ~= nil and value ~= nil then --param = value pattern matched
			value = value:gsub('"', '') --remove brackets from value
			value = value:gsub('^(%.[0-9])', '0%1') --add 0 before dot if missing at the beginning of matched string
			value = value:gsub('([^0-9])(%.[0-9])', '%10%2') --add 0 before dot if missing anywhere else
			if param:match('^font[0-9]+$') then --font declaration param matched
				local num = tonumber(param:match('font([0-9]+)'))
				if param:match('_height$') then
					if pos.font_height == nil then
						pos.font_height = {}
					end
					pos.font_height[num] = main.f_dataType(value)
				else
					value = value:lower()
					value = value:gsub('\\', '/')
					if t.font_data[value] == nil then
						if not value:match('^data/') then
							if main.f_fileExists(fileDir .. value) then
								value = fileDir .. value
							elseif main.f_fileExists('font/' .. value) then
								value = 'font/' .. value
							elseif main.f_fileExists(t.files.fight:match('^(.-)[^/\\]+$') .. value) then
								value = t.files.fight:match('^(.-)[^/\\]+$') .. value
							end
						end
						t.font_data[value] = fontNew(value)
						wait = true
					end
					if pos.font == nil then
						pos.font = {}
					end
					pos.font[num] = tostring(value)
				end
			elseif pos[param] == nil then --mugen takes into account only first occurrence
				if value:match('.+,.+') then --multiple values
					for i, c in ipairs(main.f_strsplit(',', value)) do --split value using "," delimiter
						if param:match('_anim$') then --mugen recognizes animations even if there are more values
							pos[param] = main.f_dataType(c)
							break
						elseif i == 1 then
							pos[param] = {}
							if param:match('_font$') then
								c = t.files.font[tonumber(c)]
							end
						end
						if c == nil or c == '' then
							pos[param][#pos[param] + 1] = 0
						else
							pos[param][#pos[param] + 1] = main.f_dataType(c)
						end
					end
				else --single value
					pos[param] = main.f_dataType(value)
				end
			end
		else --only valid lines left are animations
			line = line:lower()
			local value = line:match('^%s*([0-9%-]+%s*,%s*[0-9%-]+%s*,%s*[0-9%-]+%s*,%s*[0-9%-]+%s*,%s*[0-9%-]+.-)[,%s]*$') or line:match('^%s*loopstart') or line:match('^%s*interpolate offset') or line:match('^%s*interpolate angle') or line:match('^%s*interpolate scale') or line:match('^%s*interpolate blend')
			if value ~= nil then
				value = value:gsub(',%s*,', ',0,') --add missing values
				value = value:gsub(',%s*$', '')
				pos[#pos + 1] = value
			end
		end
	end
	main.loadingRefresh()
end
file:close()

--;===========================================================
--; FIX REFERENCES, LOAD DATA
--;===========================================================
local anim = ''
local facing = ''

--merge tables
motif = main.f_tableMerge(motif, t)

--general paths
local t_dir = {
	{t = {'files',            'spr'},              skip = {'^data/'},        dirs = {motif.fileDir .. motif.files.spr,                   'data/' .. motif.files.spr}},
	{t = {'files',            'snd'},              skip = {'^data/',  '^$'}, dirs = {motif.fileDir .. motif.files.snd,                   'data/' .. motif.files.snd}},
	{t = {'files',            'continue_snd'},     skip = {'^data/',  '^$'}, dirs = {motif.fileDir .. motif.files.continue_snd,          'data/' .. motif.files.continue_snd}},
	{t = {'files',            'logo_storyboard'},  skip = {'^data/',  '^$'}, dirs = {motif.fileDir .. motif.files.logo_storyboard,       'data/' .. motif.files.logo_storyboard}},
	{t = {'files',            'intro_storyboard'}, skip = {'^data/',  '^$'}, dirs = {motif.fileDir .. motif.files.intro_storyboard,      'data/' .. motif.files.intro_storyboard}},
	{t = {'files',            'select'},           skip = {'^data/'},        dirs = {motif.fileDir .. motif.files.select,                'data/' .. motif.files.select}},
	{t = {'files',            'fight'},            skip = {'^data/'},        dirs = {motif.fileDir .. motif.files.fight,                 'data/' .. motif.files.fight}},
	{t = {'music',            'title_bgm'},        skip = {'^music/', '^$'}, dirs = {motif.fileDir .. motif.music.title_bgm,             'music/' .. motif.music.title_bgm}},
	{t = {'music',            'select_bgm'},       skip = {'^music/', '^$'}, dirs = {motif.fileDir .. motif.music.select_bgm,            'music/' .. motif.music.select_bgm}},
	{t = {'music',            'vs_bgm'},           skip = {'^music/', '^$'}, dirs = {motif.fileDir .. motif.music.vs_bgm,                'music/' .. motif.music.vs_bgm}},
	{t = {'music',            'victory_bgm'},      skip = {'^music/', '^$'}, dirs = {motif.fileDir .. motif.music.victory_bgm,           'music/' .. motif.music.victory_bgm}},
	{t = {'default_ending',   'storyboard'},       skip = {'^data/',  '^$'}, dirs = {motif.fileDir .. motif.default_ending.storyboard,   'data/' .. motif.default_ending.storyboard}},
	{t = {'end_credits',      'storyboard'},       skip = {'^data/',  '^$'}, dirs = {motif.fileDir .. motif.end_credits.storyboard,      'data/' .. motif.end_credits.storyboard}},
	{t = {'game_over_screen', 'storyboard'},       skip = {'^data/',  '^$'}, dirs = {motif.fileDir .. motif.game_over_screen.storyboard, 'data/' .. motif.game_over_screen.storyboard}},
}
for i = 1, #t_dir do
	local skip = false
	for j = 1, #t_dir[i].skip do
		if motif[t_dir[i].t[1]][t_dir[i].t[2]]:match(t_dir[i].skip[j]) then
			skip = true
			break
		end
	end
	if not skip then
		for j = 1, #t_dir[i].dirs do
			if main.f_fileExists(t_dir[i].dirs[j]) then
				motif[t_dir[i].t[1]][t_dir[i].t[2]] = t_dir[i].dirs[j]
				break
			end
		end
	end
end

motif.files.spr_data = sffNew(motif.files.spr)
main.loadingRefresh()
motif.files.snd_data = sndNew(motif.files.snd)
main.loadingRefresh()
motif.files.continue_snd_data = sndNew(motif.files.continue_snd)
main.loadingRefresh()

--fadein / fadeout data
t_dir = {'title_info', 'select_info', 'vs_screen', 'victory_screen', 'win_screen', 'survival_results_screen', 'vs100kumite_results_screen', 'option_info', 'tournament_info', 'continue_screen'}
for i = 1, #t_dir do
	motif[t_dir[i]].fadein_data = main.f_fadeAnim(1, motif[t_dir[i]].fadein_time, motif[t_dir[i]].fadein_col[1], motif[t_dir[i]].fadein_col[2], motif[t_dir[i]].fadein_col[3])
	animSetWindow(motif[t_dir[i]].fadein_data, main.screenOverscan, 0, motif.info.localcoord[1], motif.info.localcoord[2])
	motif[t_dir[i]].fadeout_data = main.f_fadeAnim(0, motif[t_dir[i]].fadeout_time, motif[t_dir[i]].fadeout_col[1], motif[t_dir[i]].fadeout_col[2], motif[t_dir[i]].fadeout_col[3])
	animSetWindow(motif[t_dir[i]].fadeout_data, main.screenOverscan, 0, motif.info.localcoord[1], motif.info.localcoord[2])
end

--other entries
t_dir = {'titlebgdef', 'selectbgdef', 'versusbgdef', 'optionbgdef', 'continuebgdef', 'victorybgdef', 'resultsbgdef', 'tournamentbgdef'}
for i = 1, #t_dir do
	--ctrldef table adjustment
	for k, v in pairs(motif.ctrldef[t_dir[i]]) do
		tmp = motif.ctrldef[t_dir[i]][k].ctrl
		for j = 1, #tmp do
			--if END_TIME is omitted it should default to the same value as START_TIME
			if tmp[j].time[2] == -1 then
				tmp[j].time[2] = tmp[j].time[1]
			end
			--if LOOPTIME is omitted or set to -1, the background controller will not reset its own timer. In such case use GLOBAL_LOOPTIME
			if tmp[j].time[3] == -1 then
				tmp[j].time[3] = motif.ctrldef[t_dir[i]][k].looptime
			end
			--lowercase type name
			tmp[j].type = tmp[j].type:lower()
			--this list, if specified, overrides the default list specified in the BGCtrlDef
			if #tmp[j].ctrlid == 0 then
				for z = 1, #motif.ctrldef[t_dir[i]][k].ctrlid do
					tmp[j].ctrlid[#tmp[j].ctrlid + 1] = motif.ctrldef[t_dir[i]][k].ctrlid[z]
				end
			end
		end
	end
	--optional sff paths and data
	if motif[t_dir[i]].spr ~= '' then
		if not motif[t_dir[i]].spr:match('^data/') then
			if main.f_fileExists(motif.fileDir .. motif[t_dir[i]].spr) then
				motif[t_dir[i]].spr = motif.fileDir .. motif[t_dir[i]].spr
			elseif main.f_fileExists('data/' .. motif[t_dir[i]].spr) then
				motif[t_dir[i]].spr = 'data/' .. motif[t_dir[i]].spr
			end
		end
		motif[t_dir[i]].spr_data = sffNew(motif[t_dir[i]].spr)
		main.loadingRefresh()
	elseif motif[t_dir[i]].spr ~= 'continuebgdef' and motif[t_dir[i]].spr ~= 'tournamentbgdef' then
		motif[t_dir[i]].spr = motif.files.spr
		motif[t_dir[i]].spr_data = motif.files.spr_data
	end
	--clearcolor data
	motif[t_dir[i]].bgclearcolor_data = main.f_clearColor(motif[t_dir[i]].bgclearcolor[1], motif[t_dir[i]].bgclearcolor[2], motif[t_dir[i]].bgclearcolor[3])
	animSetWindow(motif[t_dir[i]].bgclearcolor_data, main.screenOverscan, 0, motif.info.localcoord[1], motif.info.localcoord[2])
	--background data
	motif[t_dir[i]].bg_data = {}
	local t_bgdef = motif[t_dir[i]].bg
	local prev_k = ''
	for k, v in pairs(t_bgdef) do --loop through table keys
		t_bgdef[k].type = t_bgdef[k].type:lower()
		--mugen ignores delta = 0 (defaults to 1)
		if t_bgdef[k].delta[1] == 0 then t_bgdef[k].delta[1] = 1 end
		if t_bgdef[k].delta[2] == 0 then t_bgdef[k].delta[2] = 1 end
		--add ctrl data
		motif[t_dir[i]].bg[k].ctrl = main.f_ctrlBG(t_bgdef[k], motif.ctrldef[t_dir[i]])
		--positionlink adjustment
		if t_bgdef[k].positionlink == 1 and prev_k ~= '' then
			t_bgdef[k].start[1] = t_bgdef[prev_k].start[1]
			t_bgdef[k].start[2] = t_bgdef[prev_k].start[2]
			t_bgdef[k].delta[1] = t_bgdef[prev_k].delta[1]
			t_bgdef[k].delta[2] = t_bgdef[prev_k].delta[2]
		end
		prev_k = k
		--generate anim data
		local sizeX, sizeY, offsetX, offsetY = 0, 0, 0, 0
		if t_bgdef[k].type == 'anim' then
			anim = main.f_animFromTable(motif.anim[t_bgdef[k].actionno], motif[t_dir[i]].spr_data, (t_bgdef[k].start[1] + main.normalSpriteCenter), t_bgdef[k].start[2])
		else --normal, parallax
			anim = t_bgdef[k].spriteno[1] .. ', ' .. t_bgdef[k].spriteno[2] .. ', ' .. (t_bgdef[k].start[1] + main.normalSpriteCenter) .. ', ' .. t_bgdef[k].start[2] .. ', ' .. -1
			anim = animNew(motif[t_dir[i]].spr_data, anim)
			sizeX, sizeY, offsetX, offsetY = getSpriteInfo(motif[t_dir[i]].spr, t_bgdef[k].spriteno[1], t_bgdef[k].spriteno[2])
		end
		if t_bgdef[k].trans == 'add1' then
			animSetAlpha(anim, 255, 128)
		elseif t_bgdef[k].trans == 'add' then
			animSetAlpha(anim, 255, 255)
		elseif t_bgdef[k].trans == 'sub' then
			animSetAlpha(anim, 1, 255)
		end
		animAddPos(anim, 160, 0) --for some reason needed in ikemen
		if t_bgdef[k].window ~= nil then
			animSetWindow(
				anim,
				t_bgdef[k].window[1],
				t_bgdef[k].window[2],
				t_bgdef[k].window[3] - t_bgdef[k].window[1] + 1,
				t_bgdef[k].window[4] - t_bgdef[k].window[2] + 1
			)
		else
			animSetWindow(anim, main.screenOverscan, 0, motif.info.localcoord[1], motif.info.localcoord[2])
		end
		if t_bgdef[k].tilespacing[2] == nil then t_bgdef[k].tilespacing[2] = t_bgdef[k].tilespacing[1] end
		if t_bgdef[k].type == 'parallax' then
			animSetTile(anim, t_bgdef[k].tile[1], 0, t_bgdef[k].tilespacing[1] + sizeX, t_bgdef[k].tilespacing[2] + sizeY)
		else
			animSetTile(anim, t_bgdef[k].tile[1], t_bgdef[k].tile[2], t_bgdef[k].tilespacing[1] + sizeX, t_bgdef[k].tilespacing[2] + sizeY)
		end
		if t_bgdef[k].mask == 1 or t_bgdef[k].type ~= 'normal' or (t_bgdef[k].trans ~= '' and t_bgdef[k].trans ~= 'none') then
			animSetColorKey(anim, 0)
		else
			animSetColorKey(anim, -1)
		end
		
		-- Scale non animated sprites
		animUpdate(anim)
		animSetScale(anim, 1 ,1)
		
		motif[t_dir[i]].bg_data[k] = anim
		main.loadingRefresh()
	end
end

local function f_facing(var)
	if var == -1 then
		return 'H'
	else
		return nil
	end
end

local function f_alphaToTable(var) --not used yet
	var = var:match('^%s*(.-)%s*$')
	var = var:lower()
	if var:match('^a$') then
		return {255, 255} --AS256D256
	elseif var:match('^a1$') then
		return {255, 128} --AS256D128
	elseif var:match('^s$') then
		return {1, 255} --AS0D256
	elseif var:match('^s1$') then
		return {1, 128} --are these values correct for S1?
	elseif var:match('^as[0-9]+d[0-9]+$') then
		local tabl = {}
		tabl[1] = tonumber(var:match('^as([0-9]+)'))
		tabl[2] = tonumber(var:match('d([0-9]+)$'))
		return tabl
	else
		return nil
	end
end

t = motif.select_info
t_dir = {
	{s = 'cell_bg_',                      x = 0,                                                   y = 0},
	{s = 'cell_random_',                  x = 0,                                                   y = 0},
	{s = 'p1_cursor_active_',             x = 0,                                                   y = 0},
	{s = 'p1_cursor_done_',               x = 0,                                                   y = 0},
	{s = 'p2_cursor_active_',             x = 0,                                                   y = 0},
	{s = 'p2_cursor_done_',               x = 0,                                                   y = 0},
	{s = 'p1_teammenu_bg_',               x = t.p1_teammenu_pos[1],                                y = t.p1_teammenu_pos[2]},
	{s = 'p1_teammenu_selftitle_',        x = t.p1_teammenu_pos[1],                                y = t.p1_teammenu_pos[2]},
	{s = 'p1_teammenu_enemytitle_',       x = t.p1_teammenu_pos[1],                                y = t.p1_teammenu_pos[2]},
	{s = 'p1_teammenu_item_cursor_',      x = t.p1_teammenu_pos[1] + t.p1_teammenu_item_offset[1], y = t.p1_teammenu_pos[2] + t.p1_teammenu_item_offset[2]},
	{s = 'p1_teammenu_value_icon_',       x = t.p1_teammenu_pos[1] + t.p1_teammenu_item_offset[1], y = t.p1_teammenu_pos[2] + t.p1_teammenu_item_offset[2]},
	{s = 'p1_teammenu_value_empty_icon_', x = t.p1_teammenu_pos[1] + t.p1_teammenu_item_offset[1], y = t.p1_teammenu_pos[2] + t.p1_teammenu_item_offset[2]},
	{s = 'p2_teammenu_bg_',               x = t.p2_teammenu_pos[1],                                y = t.p2_teammenu_pos[2]},
	{s = 'p2_teammenu_selftitle_',        x = t.p2_teammenu_pos[1],                                y = t.p2_teammenu_pos[2]},
	{s = 'p2_teammenu_enemytitle_',       x = t.p2_teammenu_pos[1],                                y = t.p2_teammenu_pos[2]},
	{s = 'p2_teammenu_item_cursor_',      x = t.p2_teammenu_pos[1] + t.p2_teammenu_item_offset[1], y = t.p2_teammenu_pos[2] + t.p2_teammenu_item_offset[2]},
	{s = 'p2_teammenu_value_icon_',       x = t.p2_teammenu_pos[1] + t.p2_teammenu_item_offset[1], y = t.p2_teammenu_pos[2] + t.p2_teammenu_item_offset[2]},
	{s = 'p2_teammenu_value_empty_icon_', x = t.p2_teammenu_pos[1] + t.p2_teammenu_item_offset[1], y = t.p2_teammenu_pos[2] + t.p2_teammenu_item_offset[2]},
}
for i = 1, #t_dir do
	--if i <= 2 and #t[t_dir[i].s .. 'spr'] == 0 and t[t_dir[i].s .. 'anim'] ~= nil and motif.anim[t[t_dir[i].s .. 'anim']] ~= nil then --cell_bg_, cell_random_
	--	for j = 1, #motif.anim[t[t_dir[i].s .. 'anim']] do
	--		for k, c in ipairs(main.f_strsplit(',', motif.anim[t[t_dir[i].s .. 'anim']][j])) do
	--			if c:match('loopstart') then
	--				break
	--			elseif k <= 2 then
	--				t[t_dir[i].s .. 'spr'][k] = tonumber(c)
	--			elseif k == 7 and type(c) == 'string' then
	--				t[t_dir[i].s .. 'alpha'] = f_alphaToTable(c)
	--			end
	--		end
	--	end
	--end
	if #t[t_dir[i].s .. 'spr'] > 0 then --create sprite data
		if #t[t_dir[i].s .. 'spr'] == 1 then --fix values
			if type(t[t_dir[i].s .. 'spr'][1]) == 'string' then
				t[t_dir[i].s .. 'spr'] = {tonumber(t[t_dir[i].s .. 'spr'][1]:match('^([0-9]+)')), 0}
			else
				t[t_dir[i].s .. 'spr'] = {t[t_dir[i].s .. 'spr'][1], 0}
			end
		end
		if t[t_dir[i].s .. 'facing'] == -1 then facing = ', H' else facing = '' end
		anim = t[t_dir[i].s .. 'spr'][1] .. ', ' .. t[t_dir[i].s .. 'spr'][2] .. ', ' .. t[t_dir[i].s .. 'offset'][1] + t_dir[i].x .. ', ' .. t[t_dir[i].s .. 'offset'][2] + t_dir[i].y .. ', -1' .. facing
		t[t_dir[i].s .. 'data'] = animNew(motif.selectbgdef.spr_data, anim)
		animSetScale(t[t_dir[i].s .. 'data'], t[t_dir[i].s .. 'scale'][1], t[t_dir[i].s .. 'scale'][2])
		animUpdate(t[t_dir[i].s .. 'data'])
	elseif t[t_dir[i].s .. 'anim'] ~= nil and motif.anim[t[t_dir[i].s .. 'anim']] ~= nil then --create animation data
		t[t_dir[i].s .. 'data'] = main.f_animFromTable(
			motif.anim[t[t_dir[i].s .. 'anim']],
			motif.selectbgdef.spr_data,
			t[t_dir[i].s .. 'offset'][1] + t_dir[i].x,
			t[t_dir[i].s .. 'offset'][2] + t_dir[i].y,
			t[t_dir[i].s .. 'scale'][1],
			t[t_dir[i].s .. 'scale'][2],
			f_facing(t[t_dir[i].s .. 'facing'])
		)
	else --create dummy data
		t[t_dir[i].s .. 'data'] = animNew(motif.selectbgdef.spr_data, '-1, -1, 0, 0, -1')
		animUpdate(t[t_dir[i].s .. 'data'])
	end
	animSetWindow(t[t_dir[i].s .. 'data'], main.screenOverscan, 0, motif.info.localcoord[1], motif.info.localcoord[2])
	--animAddPos(t[t_dir[i].s .. 'data'], 160, 0) --for some reason needed in ikemen (but not in this case)
	main.loadingRefresh()
end

t = motif.continue_screen
if motif.anim[t.continue_anim] ~= nil then
	t.continue_anim_data = main.f_animFromTable(
		motif.anim[t.continue_anim],
		motif.continuebgdef.spr_data,
		t.continue_offset[1],
		t.continue_offset[2],
		t.continue_scale[1],
		t.continue_scale[2]
	)
	animSetWindow(t.continue_anim_data, main.screenOverscan, 0, motif.info.localcoord[1], motif.info.localcoord[2])
end

if motif.vs_screen.p1_name_active_font == nil then
	motif.vs_screen.p1_name_active_font = {motif.vs_screen.p1_name_font[1], motif.vs_screen.p1_name_font[2], motif.vs_screen.p1_name_font[3]}
	motif.vs_screen.p1_name_active_font_scale = {motif.vs_screen.p1_name_font_scale[1], motif.vs_screen.p1_name_font_scale[2]}
end
if motif.vs_screen.p2_name_active_font == nil then
	motif.vs_screen.p2_name_active_font = {motif.vs_screen.p2_name_font[1], motif.vs_screen.p2_name_font[2], motif.vs_screen.p2_name_font[3]}
	motif.vs_screen.p2_name_active_font_scale = {motif.vs_screen.p2_name_font_scale[1], motif.vs_screen.p2_name_font_scale[2]}
end

--motif.ctrldef = nil
main.f_printTable(motif, "debug/t_motif.txt")

return motif
