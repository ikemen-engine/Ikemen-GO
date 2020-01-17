--;===========================================================
--; LOCALCOORD VALUES
--;===========================================================

local def = config.Motif
if main.flags['-r'] ~= nil then
	local case = main.flags['-r']:lower()
	if case:match('^data[/\\]') and main.f_fileExists(main.flags['-r']) then
		def = main.flags['-r']
	elseif case:match('%.def$') and main.f_fileExists('data/' .. main.flags['-r']) then
		def = 'data/' .. main.flags['-r']
	elseif main.f_fileExists('data/' .. main.flags['-r'] .. '/system.def') then
		def = 'data/' .. main.flags['-r'] .. '/system.def'
	end
end

local file = io.open(def, 'r')
local s_file = file:read("*all")
file:close()
local localX, localY = s_file:match('localcoord%s*=%s*(%d+)%s*,%s*(%d+)')
--local scaleX = 1
--local scaleY = 1
if localX ~= nil then
	localX = tonumber(localX)
	--scaleX = localX / 320
else
	localX = 320
end
if localY ~= nil then
	localY = tonumber(localY)
	--scaleY = localY / 240
else
	localY = 240
end
--local coords_fix = 0
--if scaleY > 1 then
--	coords_fix = math.floor(scaleY - 1)
--end

--;===========================================================
--; DEFAULT VALUES
--;===========================================================
--This pre-made table (3/4 of the whole file) contains all default values used in screenpack. New table from parsed DEF file is merged on top of this one.
--This is important because there are more params available in Ikemen. Whole screenpack code refers to these values.
local motif =
{
	def = def,
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
		debug_font = 'f-6x9.def', --Ikemen feature
		debug_script = 'script/debug.lua', --Ikemen feature
		font =
		{
			[1] = 'f-4x6.def',
			[2] = 'f-6x9.def',
			[3] = 'jg.fnt',
		},
		font_height = {}
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
		continue_bgm = 'sound/CONTINUE.ogg', --Ikemen feature
		continue_bgm_volume = 100, --Ikemen feature
		continue_bgm_loop = 1, --Ikemen feature
		continue_bgm_loopstart = 0, --Ikemen feature
		continue_bgm_loopend = 0, --Ikemen feature
		continue_end_bgm = 'sound/GAME_OVER.ogg', --Ikemen feature
		continue_end_bgm_volume = 100, --Ikemen feature
		continue_end_bgm_loop = 0, --Ikemen feature
		continue_end_bgm_loopstart = 0, --Ikemen feature
		continue_end_bgm_loopend = 0, --Ikemen feature
		results_bgm = '', --Ikemen feature
		results_bgm_volume = 100, --Ikemen feature
		results_bgm_loop = 1, --Ikemen feature
		results_bgm_loopstart = 0, --Ikemen feature
		results_bgm_loopend = 0, --Ikemen feature
		tournament_bgm = '', --Ikemen feature
		tournament_bgm_volume = 100, --Ikemen feature
		tournament_bgm_loop = 1, --Ikemen feature
		tournament_bgm_loopstart = 0, --Ikemen feature
		tournament_bgm_loopend = 0, --Ikemen feature
	},
	title_info =
	{
		fadein_time = 10,
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadeout_time = 10,
		fadeout_col = {0, 0, 0}, --Ikemen feature
		loading_offset = {localX - math.floor(10 * localX / 320 + 0.5), localY - 10}, --Ikemen feature (310, 230)
		loading_font = {'f-4x6.def', 7, -1, 255, 255, 255, 255, 0}, --Ikemen feature
		loading_font_scale = {1.0, 1.0}, --Ikemen feature
		loading_text = 'LOADING...', --Ikemen feature
		footer1_offset = {math.floor(2 * localX / 320 + 0.5), localY - 2}, --Ikemen feature (2, 238)
		footer1_font = {'f-4x6.def', 7, 1, 255, 255, 255, 255, 0}, --Ikemen feature
		footer1_font_scale = {1.0, 1.0}, --Ikemen feature
		footer1_text = 'I.K.E.M.E.N. GO', --Ikemen feature
		footer2_offset = {localX / 2, localY - 2}, --Ikemen feature (160, 238)
		footer2_font = {'f-4x6.def', 7, 0, 255, 255, 255, 255, 0}, --Ikemen feature
		footer2_font_scale = {1.0, 1.0}, --Ikemen feature
		footer2_text = 'Press F1 for info', --Ikemen feature
		footer3_offset = {localX - math.floor(2 * localX / 320 + 0.5), localY - 2}, --Ikemen feature (318, 238)
		footer3_font = {'f-4x6.def', 7, -1, 255, 255, 255, 255, 0}, --Ikemen feature
		footer3_font_scale = {1.0, 1.0}, --Ikemen feature
		footer3_text = 'Plus v0.9', --Ikemen feature
		footer_boxbg_visible = 1, --Ikemen feature
		footer_boxbg_coords = {0, localY - 7, localX - 1, localY - 1}, --Ikemen feature (0, 233, 319, 239)
		footer_boxbg_col = {0, 0, 64}, --Ikemen feature
		footer_boxbg_alpha = {255, 100}, --Ikemen feature
		connecting_offset = {math.floor(10 * localX / 320 + 0.5), 40}, --Ikemen feature
		connecting_font = {'f-6x9.def', 0, 1, 255, 255, 255, 255, 0}, --Ikemen feature
		connecting_font_scale = {1.0, 1.0}, --Ikemen feature
		connecting_host_text = 'Waiting for player 2... (%s)', --Ikemen feature
		connecting_join_text = 'Now connecting... (%s)', --Ikemen feature
		connecting_boxbg_coords = {0, 0, config.Width, config.Height}, --Ikemen feature (0, 0, 320, 240)
		connecting_boxbg_col = {0, 0, 0}, --Ikemen feature
		connecting_boxbg_alpha = {20, 100}, --Ikemen feature
		input_ip_name_text = 'Enter Host display name, e.g. John.\nExisting entries can be removed with DELETE button.', --Ikemen feature
		input_ip_address_text = 'Enter Host IP address, e.g. 127.0.0.1\nCopied text can be pasted with INSERT button.', --Ikemen feature
		menu_pos = {159, 158},
		menu_item_font = {'f-6x9.def', 7, 0, 255, 255, 255, 255, 0},
		menu_item_font_scale = {1.0, 1.0}, --broken parameter in mugen 1.1: http://mugenguild.com/forum/msg.1905756
		menu_item_active_font = {'f-6x9.def', 0, 0, 255, 255, 255, 255, 0},
		menu_item_active_font_scale = {1.0, 1.0}, --broken parameter in mugen 1.1: http://mugenguild.com/forum/msg.1905756
		menu_item_spacing = {0, 13},
		menu_itemname_arcade = 'ARCADE',
		menu_itemname_versus = 'VS MODE',
		menu_itemname_teamarcade = 'TEAM ARCADE',
		menu_itemname_teamversus = 'TEAM VERSUS',
		menu_itemname_online = '', --Ikemen feature (NETWORK)
		menu_itemname_teamcoop = 'TEAM CO-OP',
		menu_itemname_survival = 'SURVIVAL',
		menu_itemname_survivalcoop = 'SURVIVAL CO-OP',
		menu_itemname_storymode = '', --Ikemen feature (STORY MODE, not implemented yet)
		menu_itemname_timeattack = '', --Ikemen feature (TIME ATTACK, not implemented yet)
		menu_itemname_tournament = '', --Ikemen feature (TOURNAMENT, not implemented yet)
		menu_itemname_training = 'TRAINING',
		menu_itemname_watch = 'WATCH',
		menu_itemname_extras = '', --Ikemen feature (EXTRAS)
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
		menu_itemname_randomtest = 'DEMO', --Ikemen feature
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
		menu_boxcursor_col = {255, 255, 255}, --Ikemen feature
		menu_boxcursor_alpharange = {10, 40, 2, 255, 255, 0}, --Ikemen feature
		cursor_move_snd = {100, 0},
		cursor_done_snd = {100, 1},
		cancel_snd = {100, 2},
	},
	titlebgdef =
	{
		spr = '',
		bgclearcolor = {0, 0, 0},
	},
	infobox =
	{
		title = '', --Ikemen feature
		title_pos = {159, 19}, --Ikemen feature
		title_font = {'f-6x9.def', 0, 0, 255, 255, 255, 255, 0}, --Ikemen feature
		title_font_scale = {1.0, 1.0}, --Ikemen feature
		text = "Welcome to SUEHIRO's I.K.E.M.E.N GO engine!\n\n* This is a public development release, for testing purposes.\n* This build may contain bugs and incomplete features.\n* Your help and cooperation are appreciated!\n* I.K.E.M.E.N GO source code: https://osdn.net/users/supersuehiro/\n* Ikemen GO Plus source code: https://github.com/K4thos/Ikemen-GO-Plus", --Ikemen feature (requires new 'text = ' entry under [Infobox] section)
		text_pos = {25, 30}, --Ikemen feature
		text_font = {'f-4x6.def', 7, 1, 255, 255, 255, 255, 0},
		text_font_scale = {1.0, 1.0}, --Ikemen feature
		text_spacing = {0, 13}, --Ikemen feature
		boxbg_coords = {0, 0, config.Width, config.Height}, --Ikemen feature (0, 0, 320, 240)
		boxbg_col = {0, 0, 0}, --Ikemen feature
		boxbg_alpha = {20, 100}, --Ikemen feature
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
		cell_spacing = {2, 2}, --Ikemen feature (optionally accepts x, y values instead of a single one for both coordinates)
		--cell_bg_anim = nil,
		cell_bg_spr = {},
		cell_bg_offset = {0, 0},
		cell_bg_facing = 1,
		cell_bg_scale = {1.0, 1.0},
		--cell_random_anim = nil,
		cell_random_spr = {},
		cell_random_offset = {0, 0},
		cell_random_facing = 1,
		cell_random_scale = {1.0, 1.0},
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
		title_font = {'jg.fnt', 0, 0, 255, 255, 255, 255, 0},
		title_font_scale = {1.0, 1.0},
		title_text_arcade = 'Arcade', --Ikemen feature
		title_text_versus = 'Versus Mode', --Ikemen feature
		title_text_teamarcade = 'Team Arcade', --Ikemen feature
		title_text_teamversus = 'Team Versus', --Ikemen feature
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
		p1_face_num = 1, --Ikemen feature
		p1_face_spacing = {0, 0}, --Ikemen feature
		p1_c1_face_offset = {0, 0}, --Ikemen feature
		p1_c1_face_scale = {1.0, 1.0}, --Ikemen feature
		p1_c2_face_offset = {0, 0}, --Ikemen feature
		p1_c2_face_scale = {1.0, 1.0}, --Ikemen feature
		p1_c3_face_offset = {0, 0}, --Ikemen feature
		p1_c3_face_scale = {1.0, 1.0}, --Ikemen feature
		p1_c4_face_offset = {0, 0}, --Ikemen feature
		p1_c4_face_scale = {1.0, 1.0}, --Ikemen feature
		p2_face_spr = {9000, 1},
		p2_face_offset = {0, 0},
		p2_face_facing = -1,
		p2_face_scale = {1.0, 1.0},
		p2_face_window = {0, 0, 0, 0},
		p2_face_num = 1, --Ikemen feature
		p2_face_spacing = {0, 0}, --Ikemen feature
		p2_c1_face_offset = {0, 0}, --Ikemen feature
		p2_c1_face_scale = {1.0, 1.0}, --Ikemen feature
		p2_c2_face_offset = {0, 0}, --Ikemen feature
		p1_c2_face_scale = {1.0, 1.0}, --Ikemen feature
		p2_c3_face_offset = {0, 0}, --Ikemen feature
		p2_c3_face_scale = {1.0, 1.0}, --Ikemen feature
		p2_c4_face_offset = {0, 0}, --Ikemen feature
		p2_c4_face_scale = {1.0, 1.0}, --Ikemen feature
		p1_name_offset = {0, 0},
		p1_name_font = {'jg.fnt', 4, 1, 255, 255, 255, 255, 0},
		p1_name_font_scale = {1.0, 1.0},
		p1_name_spacing = {0, 14},
		p2_name_offset = {0, 0},
		p2_name_font = {'jg.fnt', 1, -1, 255, 255, 255, 255, 0},
		p2_name_font_scale = {1.0, 1.0},
		p2_name_spacing = {0, 14},
		stage_pos = {0, 0},
		stage_active_font = {'f-4x6.def', 0, 0, 255, 255, 255, 255, 0},
		stage_active_font_scale = {1.0, 1.0},
		stage_active2_font = {'f-4x6.def', 0, 0, 255, 255, 255, 255, 0},
		stage_active2_font_scale = {1.0, 1.0},
		stage_done_font = {'f-4x6.def', 0, 0, 255, 255, 255, 255, 0},
		stage_done_font_scale = {1.0, 1.0},
		stage_text = 'Stage %i: %s', --Ikemen feature
		stage_random_text = 'Stage: Random', --Ikemen feature
		stage_text_spacing = {0, 14}, --Ikemen feature
		stage_portrait_spr = {9000, 0}, --Ikemen feature
		stage_portrait_offset = {0, 0}, --Ikemen feature
		stage_portrait_scale = {1.0, 1.0}, --Ikemen feature
		stage_portrait_random_spr = {}, --Ikemen feature
		--stage_portrait_random_anim = nil, --Ikemen feature
		stage_portrait_random_offset = {0, 0}, --Ikemen feature
		stage_portrait_random_scale = {1.0, 1.0}, --Ikemen feature
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
		p1_teammenu_selftitle_font = {'jg.fnt', 0, 1, 255, 255, 255, 255, 0},
		p1_teammenu_selftitle_font_scale = {1.0, 1.0},
		p1_teammenu_selftitle_text = '',
		--p1_teammenu_enemytitle_anim = nil,
		p1_teammenu_enemytitle_spr = {},
		p1_teammenu_enemytitle_offset = {0, 0},
		p1_teammenu_enemytitle_facing = 1,
		p1_teammenu_enemytitle_scale = {1.0, 1.0},
		p1_teammenu_enemytitle_font = {'jg.fnt', 0, 1, 255, 255, 255, 255, 0},
		p1_teammenu_enemytitle_font_scale = {1.0, 1.0},
		p1_teammenu_enemytitle_text = '',
		p1_teammenu_move_snd = {100, 0},
		p1_teammenu_value_snd = {100, 0},
		p1_teammenu_done_snd = {100, 1},
		p1_teammenu_item_offset = {0, 0},
		p1_teammenu_item_spacing = {0, 15},
		p1_teammenu_item_font_offset = {0, 0}, --Ikemen feature
		p1_teammenu_item_font = {'jg.fnt', 0, 1, 255, 255, 255, 255, 0},
		p1_teammenu_item_font_scale = {1.0, 1.0},
		p1_teammenu_item_active_font = {'jg.fnt', 3, 1, 255, 255, 255, 255, 0},
		p1_teammenu_item_active_font_scale = {1.0, 1.0},
		p1_teammenu_item_active2_font = {'jg.fnt', 0, 1, 255, 255, 255, 255, 0},
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
		p2_teammenu_selftitle_font = {'jg.fnt', 0, -1, 255, 255, 255, 255, 0},
		p2_teammenu_selftitle_font_scale = {1.0, 1.0},
		p2_teammenu_selftitle_text = '',
		--p2_teammenu_enemytitle_anim = nil,
		p2_teammenu_enemytitle_spr = {},
		p2_teammenu_enemytitle_offset = {0, 0},
		p2_teammenu_enemytitle_facing = 1,
		p2_teammenu_enemytitle_scale = {1.0, 1.0},
		p2_teammenu_enemytitle_font = {'jg.fnt', 0, -1, 255, 255, 255, 255, 0},
		p2_teammenu_enemytitle_font_scale = {1.0, 1.0},
		p2_teammenu_enemytitle_text = '',
		p2_teammenu_move_snd = {100, 0},
		p2_teammenu_value_snd = {100, 0},
		p2_teammenu_done_snd = {100, 1},
		p2_teammenu_item_offset = {0, 0},
		p2_teammenu_item_spacing = {0, 15},
		p2_teammenu_item_font_offset = {0, 0}, --Ikemen feature
		p2_teammenu_item_font = {'jg.fnt', 0, -1, 255, 255, 255, 255, 0},
		p2_teammenu_item_font_scale = {1.0, 1.0},
		p2_teammenu_item_active_font = {'jg.fnt', 1, -1, 255, 255, 255, 255, 0},
		p2_teammenu_item_active_font_scale = {1.0, 1.0},
		p2_teammenu_item_active2_font = {'jg.fnt', 0, -1, 255, 255, 255, 255, 0},
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
		timer_enabled = 0, --Ikemen feature
		timer_offset = {159, 39}, --Ikemen feature
		timer_font = {'f-6x9.def', 0, 0, 255, 255, 255, 255, 0}, --Ikemen feature
		timer_font_scale = {1.0, 1.0}, --Ikemen feature
		timer_count = 99, --Ikemen feature
		timer_framespercount = 60, --Ikemen feature
		timer_displaytime = 10, --Ikemen feature
	},
	selectbgdef =
	{
		spr = '',
		bgclearcolor = {0, 0, 0},
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
		match_font = {'jg.fnt', 0, 0, 255, 255, 255, 255, 0},
		match_font_scale = {1.0, 1.0},
		p1_pos = {0, 0},
		p1_spr = {9000, 1},
		p1_offset = {0, 0},
		p1_facing = 1,
		p1_scale = {1.0, 1.0},
		--p1_window = {0, 0, 0, 0}, --not implemented yet
		p1_num = 1, --Ikemen feature
		p1_spacing = {0, 0}, --Ikemen feature
		p1_c1_offset = {0, 0}, --Ikemen feature
		p1_c1_scale = {1.0, 1.0}, --Ikemen feature
		p1_c2_offset = {0, 0}, --Ikemen feature
		p1_c2_scale = {1.0, 1.0}, --Ikemen feature
		p1_c3_offset = {0, 0}, --Ikemen feature
		p1_c3_scale = {1.0, 1.0}, --Ikemen feature
		p1_c4_offset = {0, 0}, --Ikemen feature
		p1_c4_scale = {1.0, 1.0}, --Ikemen feature
		p2_pos = {0, 0},
		p2_spr = {9000, 1}, --not used in Ikemen (same as p1_spr)
		p2_offset = {0, 0},
		p2_facing = -1,
		p2_scale = {1.0, 1.0},
		--p2_window = {0, 0, 0, 0}, --not implemented yet
		p2_num = 1, --Ikemen feature
		p2_spacing = {0, 0}, --Ikemen feature
		p2_c1_offset = {0, 0}, --Ikemen feature
		p2_c1_scale = {1.0, 1.0}, --Ikemen feature
		p2_c2_offset = {0, 0}, --Ikemen feature
		p2_c2_scale = {1.0, 1.0}, --Ikemen feature
		p2_c3_offset = {0, 0}, --Ikemen feature
		p2_c3_scale = {1.0, 1.0}, --Ikemen feature
		p2_c4_offset = {0, 0}, --Ikemen feature
		p2_c4_scale = {1.0, 1.0}, --Ikemen feature
		p1_name_pos = {0, 0},
		p1_name_offset = {0, 0},
		p1_name_font = {'jg.fnt', 0, 0, 255, 255, 255, 255, 0},
		p1_name_font_scale = {1.0, 1.0},
		p1_name_spacing = {0, 14},
		p2_name_pos = {0, 0},
		p2_name_offset = {0, 0},
		p2_name_font = {'jg.fnt', 0, 0, 255, 255, 255, 255, 0},
		p2_name_font_scale = {1.0, 1.0},
		p2_name_spacing = {0, 14},
		--p1_name_active_font = {'jg.fnt', 0, 0, 255, 255, 255, 255, 0}, --Ikemen feature
		--p1_name_active_font_scale = {1.0, 1.0}, --Ikemen feature
		--p2_name_active_font = {'jg.fnt', 0, 0, 255, 255, 255, 255, 0}, --Ikemen feature
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
	},
	demo_mode =
	{
		enabled = 1,
		select_enabled = 0, --not supported yet
		vsscreen_enabled = 0, --not supported yet
		title_waittime = 600,
		fight_endtime = 1500,
		fight_playbgm = 0,
		fight_stopbgm = 0,
		fight_bars_display = 0,
		intro_waitcycles = 1,
		debuginfo = 0,
		fadein_time = 50, --Ikemen feature
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadeout_time = 50, --Ikemen feature
		fadeout_col = {0, 0, 0}, --Ikemen feature
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
		credits_font = {'jg.fnt', 0, 1, 255, 255, 255, 255, 0}, --Ikemen feature
		credits_font_scale = {1.0, 1.0}, --Ikemen feature
		--enabled = 1, --not used in Ikemen
		--pos = {160, 240}, --not used in Ikemen
		--continue_text = 'CONTINUE?', --not used in Ikemen
		--continue_font = {'f-4x6.def', 0, 0, 255, 255, 255, 255, 0}, --not used in Ikemen
		--continue_font_scale = {1.0, 1.0}, --not used in Ikemen
		--continue_offset = {0, 0}, --not used in Ikemen
		--yes_text = 'YES', --not used in Ikemen
		--yes_font = {'f-4x6.def', 0, 0, 128, 128, 128}, --not used in Ikemen
		--yes_font_scale = {1.0, 1.0}, --not used in Ikemen
		--yes_offset = {-80, 60}, --not used in Ikemen
		--yes_active_text = 'YES', --not used in Ikemen
		--yes_active_font = {'f-4x6.def', 3, 0, 255, 255, 255, 255, 0}, --not used in Ikemen
		--yes_active_font_scale = {1.0, 1.0}, --not used in Ikemen
		--yes_active_offset = {-80, 60}, --not used in Ikemen
		--no_text = 'NO', --not used in Ikemen
		--no_font = {'f-4x6.def', 0, 0, 128, 128, 128}, --not used in Ikemen
		--no_font_scale = {1.0, 1.0}, --not used in Ikemen
		--no_offset = {80, 60}, --not used in Ikemen
		--no_active_text = 'NO', --not used in Ikemen
		--no_active_font = {'f-4x6.def', 3, 0, 255, 255, 255, 255, 0}, --not used in Ikemen
		--no_active_font_scale = {1.0, 1.0}, --not used in Ikemen
		--no_active_offset = {80, 60}, --not used in Ikemen
	},
	continuebgdef =
	{
		spr = 'data/continue.sff', --Ikemen feature
		bgclearcolor = {0, 0, 0}, --Ikemen feature
	},
	game_over_screen =
	{
		enabled = 1,
		storyboard = '',
	},
	victory_screen =
	{
		enabled = 0,
		cpu_enabled = 1, --Ikemen feature
		vs_enabled = 1, --Ikemen feature
		looser_name_enabled = 0, --Ikemen feature
		winner_teamko_enabled = 0, --Ikemen feature
		time = 300,
		fadein_time = 8,
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadeout_time = 15,
		fadeout_col = {0, 0, 0}, --Ikemen feature
		p1_pos = {0, 0},
		p1_spr = {9000, 2},
		p1_offset = {100, 20},
		p1_facing = 1,
		p1_scale = {1.0, 1.0},
		--p1_window = {0, 0, 319, 160}, --not implemented yet
		p1_num = 1, --Ikemen feature
		p1_c1_offset = {0, 0}, --Ikemen feature
		p1_c1_scale = {1.0, 1.0}, --Ikemen feature
		p1_c2_offset = {0, 0}, --Ikemen feature
		p1_c2_scale = {1.0, 1.0}, --Ikemen feature
		p1_c3_offset = {0, 0}, --Ikemen feature
		p1_c3_scale = {1.0, 1.0}, --Ikemen feature
		p1_c4_offset = {0, 0}, --Ikemen feature
		p1_c4_scale = {1.0, 1.0}, --Ikemen feature
		p1_name_offset = {20, 180},
		p1_name_font = {'jg.fnt', 0, 1, 255, 255, 255, 255, 0},
		p1_name_font_scale = {1.0, 1.0},
		p2_pos = {0, 0}, --Ikemen feature
		p2_offset = {100, 20}, --Ikemen feature
		p2_facing = 1, --Ikemen feature
		p2_scale = {1.0, 1.0}, --Ikemen feature
		--p2_window = {0, 0, 319, 160}, --Ikemen feature (not implemented yet)
		p2_num = 0, --Ikemen feature
		p2_c1_offset = {0, 0}, --Ikemen feature
		p2_c1_scale = {1.0, 1.0}, --Ikemen feature
		p2_c2_offset = {0, 0}, --Ikemen feature
		p2_c2_scale = {1.0, 1.0}, --Ikemen feature
		p2_c3_offset = {0, 0}, --Ikemen feature
		p2_c3_scale = {1.0, 1.0}, --Ikemen feature
		p2_c4_offset = {0, 0}, --Ikemen feature
		p2_c4_scale = {1.0, 1.0}, --Ikemen feature
		p2_name_offset = {20, 180}, --Ikemen feature
		p2_name_font = {'jg.fnt', 0, 1, 255, 255, 255, 255, 0}, --Ikemen feature
		p2_name_font_scale = {1.0, 1.0}, --Ikemen feature
		winquote_text = 'Winner!',
		winquote_offset = {20, 192},
		winquote_font = {'f-6x9.def', 0, 1, 255, 255, 255, 255, 0},
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
		wintext_font = {'f-6x9.def', 0, 0, 255, 255, 255, 255, 0},
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
		winstext_font = {'jg.fnt', 0, 0, 255, 255, 255, 255, 0},
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
		winstext_font = {'jg.fnt', 0, 0, 255, 255, 255, 255, 0}, --Ikemen feature
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
	},
	option_info =
	{
		fadein_time = 10, --check winmugen values
		fadein_col = {0, 0, 0}, --Ikemen feature
		fadeout_time = 10, --check winmugen values
		fadeout_col = {0, 0, 0}, --Ikemen feature
		title_offset = {159, 19},
		title_font = {'f-6x9.def', 0, 0, 255, 255, 255, 255, 0},
		title_font_scale = {1.0, 1.0},
		title_text_main = 'OPTIONS', --Ikemen feature
		title_text_arcade = 'ARCADE SETTINGS', --Ikemen feature
		title_text_gameplay = 'GAMEPLAY SETTINGS', --Ikemen feature
		title_text_advgameplay = 'ADVANCED SETTINGS', --Ikemen feature
		title_text_video = 'VIDEO SETTINGS', --Ikemen feature
		title_text_res = 'RESOLUTION SETTINGS', --Ikemen feature
		title_text_externalshaders = 'SHADER SETTINGS', --Ikemen feature
		title_text_audio = 'AUDIO SETTINGS', --Ikemen feature
		title_text_engine = 'ENGINE SETTINGS', --Ikemen feature
		title_text_input = 'INPUT SETTINGS', --Ikemen feature
		title_text_key = 'KEY SETTINGS', --Ikemen feature
		title_text_controller = 'CONTROLLER SETTINGS', --Ikemen feature
		menu_pos = {85, 33}, --Ikemen feature
		menu_item_font = {'f-6x9.def', 7, 1, 255, 255, 255, 255, 0}, --Ikemen feature
		menu_item_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_active_font = {'f-6x9.def', 0, 1, 255, 255, 255, 255, 0}, --Ikemen feature
		menu_item_active_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_selected_font = {'f-6x9.def', 4, 1, 255, 255, 255, 255, 0}, --Ikemen feature
		menu_item_selected_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_selected_active_font = {'f-6x9.def', 4, 1, 255, 255, 255, 255, 0}, --Ikemen feature
		menu_item_selected_active_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_value_font = {'f-6x9.def', 7, -1, 255, 255, 255, 255, 0}, --Ikemen feature
		menu_item_value_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_value_active_font = {'f-6x9.def', 0, -1, 255, 255, 255, 255, 0}, --Ikemen feature
		menu_item_value_active_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_value_conflict_font = {'f-6x9.def', 1, -1, 255, 255, 255, 255, 0}, --Ikemen feature
		menu_item_value_conflict_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_spacing = {150, 13}, --Ikemen feature
		menu_window_visibleitems = 16, --Ikemen feature
		menu_boxcursor_visible = 1, --Ikemen feature
		menu_boxcursor_coords = {-5, -10, 154, 2}, --Ikemen feature
		menu_boxcursor_col = {255, 255, 255}, --Ikemen feature
		menu_boxcursor_alpharange = {10, 40, 2, 255, 255, 0}, --Ikemen feature
		menu_boxbg_visible = 1, --Ikemen feature
		menu_boxbg_col = {0, 0, 0}, --Ikemen feature
		menu_boxbg_alpha = {20, 100}, --Ikemen feature
		menu_itemname_main_arcade = 'Arcade Settings', --Ikemen feature
		menu_itemname_main_gameplay = 'Gameplay Settings', --Ikemen feature
		menu_itemname_main_engine = 'Engine Settings', --Ikemen feature
		menu_itemname_main_video = 'Video Settings', --Ikemen feature
		menu_itemname_main_audio = 'Audio Settings', --Ikemen feature
		menu_itemname_main_input = 'Input Settings', --Ikemen feature
		menu_itemname_main_port = 'Port Change', --Ikemen feature
		menu_itemname_main_default = 'Default Values', --Ikemen feature
		menu_itemname_main_save = 'Save and Return', --Ikemen feature
		menu_itemname_main_back = 'Return Without Saving', --Ikemen feature
		menu_itemname_arcade_roundtime = 'Time Limit', --Ikemen feature
		menu_itemname_arcade_roundtime_none = 'None', --Ikemen feature
		menu_itemname_arcade_roundsnumsingle = 'Rounds to Win Single', --Ikemen feature
		menu_itemname_arcade_roundsnumteam = 'Rounds to Win Simul/Tag', --Ikemen feature
		menu_itemname_arcade_maxdrawgames = 'Max Draw Games', --Ikemen feature
		menu_itemname_arcade_difficulty = 'Difficulty level', --Ikemen feature
		menu_itemname_arcade_credits = 'Credits', --Ikemen feature
		menu_itemname_arcade_charchange = 'Char change at Continue', --Ikemen feature
		menu_itemname_arcade_airamping = 'AI Ramping', --Ikemen feature
		menu_itemname_arcade_aipalette = 'AI Palette', --Ikemen feature
		menu_itemname_arcade_aipalette_random = 'Random', --Ikemen feature
		menu_itemname_arcade_aipalette_default = 'Default', --Ikemen feature
		menu_itemname_arcade_back = 'Back', --Ikemen feature
		menu_itemname_video_resolution = 'Resolution', --Ikemen feature
		menu_itemname_video_fullscreen = 'Fullscreen', --Ikemen feature
		menu_itemname_video_msaa = 'MSAA', --Ikemen feature
		menu_itemname_video_externalshaders = 'Shaders', --Ikemen feature
		menu_itemname_video_back = 'Back', --Ikemen feature
		menu_itemname_video_res_320x240 = '320x240    (4:3 QVGA)', --Ikemen feature
		menu_itemname_video_res_640x480 = '640x480    (4:3 VGA)', --Ikemen feature
		menu_itemname_video_res_1280x960 = '1280x960   (4:3 Quad-VGA)', --Ikemen feature
		menu_itemname_video_res_1600x1200 = '1600x1200  (4:3 UXGA)', --Ikemen feature
		menu_itemname_video_res_960x720 = '960x720    (4:3 HD)', --Ikemen feature
		menu_itemname_video_res_1280x720 = '1280x720   (16:9 HD)', --Ikemen feature
		menu_itemname_video_res_1600x900 = '1600x900   (16:9 HD+)', --Ikemen feature
		menu_itemname_video_res_1920x1080 = '1920x1080  (16:9 FHD)', --Ikemen feature
		menu_itemname_video_res_2560x1440 = '2560x1440  (16:9 2K)', --Ikemen feature
		menu_itemname_video_res_3840x2160 = '3840x2160  (16:9 4K)', --Ikemen feature
		menu_itemname_video_res_custom = 'Custom', --Ikemen feature
		menu_itemname_video_res_back = 'Back', --Ikemen feature
		menu_itemname_video_externalshaders_disableall = 'Disable all', --Ikemen feature
		menu_itemname_video_externalshaders_back = 'Back', --Ikemen feature
		menu_itemname_audio_mastervolume = 'Master Volume', --Ikemen feature
		menu_itemname_audio_bgmvolume = 'BGM Volume', --Ikemen feature
		menu_itemname_audio_sfxvolume = 'SFX Volume', --Ikemen feature
		menu_itemname_audio_audioducking = 'Audio Ducking', --Ikemen feature
		menu_itemname_audio_back = 'Back', --Ikemen feature
		menu_itemname_gameplay_lifemul = 'Life', --Ikemen feature
		menu_itemname_gameplay_autoguard = 'Auto-Guard', --Ikemen feature
		menu_itemname_gameplay_attackpowermul = 'Attack.LifeToPowerMul', --Ikemen feature
		menu_itemname_gameplay_gethitpowermul = 'GetHit.LifeToPowerMul', --Ikemen feature
		menu_itemname_gameplay_superdefencemul = 'Super.TargetDefenceMul', --Ikemen feature
		menu_itemname_gameplay_team1vs2life = '1P Vs Team Life', --Ikemen feature
		menu_itemname_gameplay_turnsrecoverybase = 'Turns Recovery Base', --Ikemen feature
		menu_itemname_gameplay_turnsrecoverybonus = 'Turns Recovery Bonus', --Ikemen feature
		menu_itemname_gameplay_teampowershare = 'Team Power Share', --Ikemen feature
		menu_itemname_gameplay_teamlifeshare = 'Team Life Share', --Ikemen feature
		menu_itemname_gameplay_singlemode = 'Single Mode', --Ikemen feature
		menu_itemname_gameplay_numturns = 'Turns Limit', --Ikemen feature
		menu_itemname_gameplay_numsimul = 'Simul Limit', --Ikemen feature
		menu_itemname_gameplay_numtag = 'Tag Limit', --Ikemen features
		menu_itemname_gameplay_advanced = 'Advanced Settings', --Ikemen feature
		menu_itemname_gameplay_back = 'Back', --Ikemen feature
		menu_itemname_engine_allowdebugkeys = 'Debug Keys', --Ikemen feature
		menu_itemname_engine_quicklaunch = 'Quick Launch', --Ikemen feature
		menu_itemname_engine_simulmode = 'Legacy Tag Mode', --Ikemen feature
		menu_itemname_engine_helpermax = 'HelperMax', --Ikemen feature
		menu_itemname_engine_playerprojectilemax = 'PlayerProjectileMax', --Ikemen feature
		menu_itemname_engine_explodmax = 'ExplodMax', --Ikemen feature
		menu_itemname_engine_afterimagemax = 'AfterImageMax', --Ikemen feature
		menu_itemname_engine_zoomactive = 'Zoom Active', --Ikemen feature
		menu_itemname_engine_maxzoomout = 'Default Max Zoom Out', --Ikemen feature
		menu_itemname_engine_maxzoomin = 'Default Max Zoom In', --Ikemen feature
		menu_itemname_engine_zoomspeed = 'Default Zoom Speed', --Ikemen feature
		menu_itemname_engine_lifebarfontscale = 'Lifebar Font Scale', --Ikemen feature
		menu_itemname_input_keyboard = 'Key Config', --Ikemen feature
		menu_itemname_input_gamepad = 'Joystick Config', --Ikemen feature
		menu_itemname_input_system = 'System Keys', --Ikemen feature (not used yet)
		menu_itemname_input_default = 'Default Values', --Ikemen feature
		menu_itemname_input_back = 'Back', --Ikemen feature
		menu_itemname_key_p1 = 'PLAYER 1', --Ikemen feature
		menu_itemname_key_p2 = 'PLAYER 2', --Ikemen feature
		menu_itemname_key_all = 'Config all', --Ikemen feature
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
		menu_itemname_key_d = 'D', --Ikemen feature
		menu_itemname_key_w = 'W', --Ikemen feature
		menu_itemname_key_back = 'Back', --Ikemen feature
		menu_item_key_p1_font = {'f-6x9.def', 4, 0, 255, 255, 255, 255, 0}, --Ikemen feature
		menu_item_key_p1_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_key_p2_font = {'f-6x9.def', 1, 0, 255, 255, 255, 255, 0}, --Ikemen feature
		menu_item_key_p2_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_info_font = {'f-6x9.def', 5, -1, 255, 255, 255, 255, 0}, --Ikemen feature
		menu_item_info_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_item_info_active_font = {'f-6x9.def', 5, -1, 255, 255, 255, 255, 0}, --Ikemen feature
		menu_item_info_active_font_scale = {1.0, 1.0}, --Ikemen feature
		menu_itemname_info_f1 = '(F1)', --Ikemen feature
		menu_itemname_info_f2 = '(F2)', --Ikemen feature
		menu_itemname_info_esc = '(Esc)', --Ikemen feature
		menu_itemname_info_disable = 'Not used', --Ikemen feature
		menu_item_p1_pos = {91, 33}, --Ikemen feature
		menu_item_p2_pos = {230, 33}, --Ikemen feature
		menu_key_p1_pos = {39, 33}, --Ikemen feature
		menu_key_p2_pos = {178, 33}, --Ikemen feature
		menu_key_item_spacing = {101, 13}, --Ikemen feature
		menu_key_boxcursor_coords = {-5, -10, 106, 2}, --Ikemen feature
		menu_itemname_yes = 'Yes', --Ikemen feature
		menu_itemname_no = 'No', --Ikemen feature
		menu_itemname_enabled = 'Enabled', --Ikemen feature
		menu_itemname_disabled = 'Disabled', --Ikemen feature
		input_text_port = 'Type in Host Port, e.g. 7500.\nPress ENTER to accept.\nPress ESC to cancel.', --Ikemen feature
		input_text_reswidth = 'Type in screen width.\nPress ENTER to accept.\nPress ESC to cancel.', --Ikemen feature
		input_text_resheight = 'Type in screen height.\nPress ENTER to accept.\nPress ESC to cancel.', --Ikemen feature
		input_text_key = 'Press a key to assign to entry.\nPress SPACE to disable key.\nPress ESC to cancel.', --Ikemen feature
		cursor_move_snd = {100, 0},
		cursor_done_snd = {100, 1},
		cancel_snd = {100, 2},
	},
	optionbgdef =
	{
		spr = '',
		bgclearcolor = {0, 0, 0},
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
	},
	warning_info =
	{
		title = 'WARNING', --Ikemen feature
		title_pos = {159, 19}, --Ikemen feature
		title_font = {'f-6x9.def', 0, 0, 255, 255, 255, 255, 0}, --Ikemen feature
		title_font_scale = {1.0, 1.0}, --Ikemen feature
		text_chars = 'No characters in select.def available for random selection.\nPress any key to exit the program.', --Ikemen feature'
		text_stages = 'No stages in select.def available for random selection.\nPress any key to exit the program.', --Ikemen feature
		text_order = "Incorrect 'maxmatches' settings detected.\nCheck orders in [Characters] and [Options] sections\nto ensure that at least one battle is possible.\nPress any key to exit the program.", --Ikemen feature
		text_training = "Training character ('chars/Training/Training.def') not found.\nPress any key to exit the program.", --Ikemen feature
		text_reload = 'Some selected options require Ikemen to be restarted.\nPress any key to exit the program.', --Ikemen feature
		text_noreload = 'Some selected options require Ikemen to be restarted.\nPress any key to continue.', --Ikemen feature
		text_res = 'Non 4:3 resolutions require stages coded for different\naspect ratio. Change it back to 4:3 if stages look off.', --Ikemen feature
		text_keys = 'Conflict between button keys detected.\nAll keys should have unique assignment.\nFix the problem before exitting key settings.', --Ikemen feature
		text_pad = 'Controller not detected.\nCheck if your controller is plugged in.', --Ikemen feature
		text_simul = 'This is a legacy option that works only if screenpack \nhas not been updated to support both Tag and Simul \nmode selection in select screen.', --Ikemen feature
		text_pos = {25, 33}, --Ikemen feature
		text_font = {'f-6x9.def', 0, 1, 255, 255, 255, 255, 0}, --Ikemen feature
		text_font_scale = {1.0, 1.0}, --Ikemen feature
		text_spacing = {0, 13}, --Ikemen feature
		boxbg_coords = {0, 0, config.Width, config.Height}, --Ikemen feature (0, 0, 320, 240)
		boxbg_col = {0, 0, 0}, --Ikemen feature
		boxbg_alpha = {20, 100}, --Ikemen feature
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
}

--;===========================================================
--; PARSE SCREENPACK
--;===========================================================
--here starts proper screenpack DEF file parsing
main.t_sort = {}
local t = {}
local pos = t
local pos_sort = main.t_sort
local def_pos = motif
t.anim = {}
t.font_data = {['f-4x6.def'] = fontNew('f-4x6.def'), ['f-6x9.def'] = fontNew('f-6x9.def'), ['jg.fnt'] = fontNew('jg.fnt')}
local fileDir, fileName = motif.def:match('^(.-)([^/\\]+)$')
t.fileDir = fileDir
t.fileName = fileName
local tmp = ''
file = io.open(motif.def, 'r')
for line in file:lines() do
	line = line:gsub('%s*;.*$', '')
	if line:match('^%s*%[.-%s*%]%s*$') then --matched [] group
		line = line:match('^%s*%[(.-)%s*%]%s*$') --match text between []
		line = line:gsub('[%. ]', '_') --change . and space to _
		local row = tostring(line:lower())
		if row:match('^begin_action_[0-9]+$') then --matched anim
			row = tonumber(row:match('^begin_action_([0-9]+)$'))
			t.anim[row] = {}
			pos = t.anim[row]
		else --matched other []
			t[row] = {}
			main.t_sort[row] = {}
			pos = t[row]
			pos_sort = main.t_sort[row]
			def_pos = motif[row]
		end
	else --matched non [] line
		local param, value = line:match('^%s*([^=]-)%s*=%s*(.-)%s*$')
		if param ~= nil then
			param = param:gsub('[%. ]', '_') --change param . and space to _
			param = param:lower() --lowercase param
			if value ~= nil and def_pos ~= nil then --let's check if it's even a valid param
				if value == '' and (type(def_pos[param]) == 'number' or type(def_pos[param]) == 'table') then --text should remain empty
					value = '0'
				end
			end
		end
		if param ~= nil and value ~= nil then --param = value pattern matched
			value = value:gsub('"', '') --remove brackets from value
			value = value:gsub('^(%.[0-9])', '0%1') --add 0 before dot if missing at the beginning of matched string
			value = value:gsub('([^0-9])(%.[0-9])', '%10%2') --add 0 before dot if missing anywhere else
			if param:match('^font[0-9]+') then --font declaration param matched
				local num = tonumber(param:match('font([0-9]+)'))
				if param:match('_height$') then
					if pos.font_height == nil then
						pos.font_height = {}
					end
					pos.font_height[num] = main.f_dataType(value)
				else
					value = value:gsub('\\', '/')
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
								if t.files ~= nil and t.files.font ~= nil and t.files.font[tonumber(c)] ~= nil then --in case font is used before it's declared in DEF file
									c = t.files.font[tonumber(c)]
								else
									break --use default font values
								end
							end
						end
						if c == nil or c == '' then
							table.insert(pos[param], 0)
						else
							table.insert(pos[param], main.f_dataType(c))
						end
					end
				else --single value
					if param:match('_itemname_') then
						table.insert(pos_sort, param:match('_itemname_(.+)$'))
					end
					pos[param] = main.f_dataType(value)
				end
			end
		else --only valid lines left are animations
			line = line:lower()
			local value = line:match('^%s*([0-9%-]+%s*,%s*[0-9%-]+%s*,%s*[0-9%-]+%s*,%s*[0-9%-]+%s*,%s*[0-9%-]+.-)[,%s]*$') or line:match('^%s*loopstart') or line:match('%s*interpolate [oasb][fncl][fgae][sln][ed]t?')
			if value ~= nil then
				value = value:gsub(',%s*,', ',0,') --add missing values
				value = value:gsub(',%s*$', '')
				table.insert(pos, value)
			end
		end
	end
	main.loadingRefresh()
end
file:close()

--;===========================================================
--; FIX REFERENCES, LOAD DATA
--;===========================================================
--adopt old DEF code to Ikemen features
if type(t.select_info.cell_spacing) ~= "table" then
	t.select_info.cell_spacing = {t.select_info.cell_spacing, t.select_info.cell_spacing}
end

--disable scaling if element should use default values (non-existing in mugen)
motif.defaultContinue = t.continue_screen == nil or t.continue_screen.continue_anim == nil or t.continue_screen.continue_anim == motif.continue_screen.continue_anim
motif.defaultWarning = true--t.warning_info == nil or t.warning_info.text_font == nil or t.warning_info.text_font[1] == motif.warning_info.text_font[1]
motif.defaultOptions = t.option_info == nil or t.option_info.menu_item_font == nil or t.option_info.menu_item_font[1] == motif.option_info.menu_item_font[1]
motif.defaultConnecting = t.title_info == nil or t.title_info.connecting_font == nil or t.title_info.connecting_font[1] == motif.title_info.connecting_font[1]
motif.defaultInfobox = t.infobox == nil or t.infobox.text_font == nil or t.infobox.text_font[1] == motif.infobox.text_font[1]
motif.defaultLoading = false --t.title_info == nil or t.title_info.loading_font == nil or t.title_info.loading_font[1] == motif.title_info.loading_font[1]
motif.defaultFooter = false --t.title_info == nil or t.title_info.footer1_font == nil or t.title_info.footer1_font[1] == motif.title_info.footer1_font[1]
motif.defaultLocalcoord = localX == 320 and localY == 240

--merge tables
motif = main.f_tableMerge(motif, t)

--fix missing params
if motif.victory_screen.enabled == 0 then
	motif.victory_screen.cpu_enabled = 0
	motif.victory_screen.vs_enabled = 0
end

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

--data
local anim = ''
local facing = ''
t_dir = {'titlebgdef', 'selectbgdef', 'versusbgdef', 'optionbgdef', 'continuebgdef', 'victorybgdef', 'resultsbgdef', 'tournamentbgdef'}
for i = 1, #t_dir do
	--optional sff paths and data
	if motif[t_dir[i]].spr ~= '' then
		if not motif[t_dir[i]].spr:match('^data/') then
			if main.f_fileExists(motif.fileDir .. motif[t_dir[i]].spr) then
				motif[t_dir[i]].spr = motif.fileDir .. motif[t_dir[i]].spr
			elseif main.f_fileExists('data/' .. motif[t_dir[i]].spr) then
				motif[t_dir[i]].spr = 'data/' .. motif[t_dir[i]].spr
			end
		end
		motif[t_dir[i]].spr_data = sffNew(motif[t_dir[i]].spr) --does sff work with all data or just backgrounds? If the latter then it's not needed
		main.loadingRefresh()
	elseif motif[t_dir[i]].spr ~= 'continuebgdef' and motif[t_dir[i]].spr ~= 'tournamentbgdef' then
		motif[t_dir[i]].spr = motif.files.spr
		motif[t_dir[i]].spr_data = motif.files.spr_data
	end
	--backgrounds
	motif[t_dir[i]].bg = bgNew(motif.def, t_dir[i]:match('^(.+)def$'), motif[t_dir[i]].spr)
	main.loadingRefresh()
end

for k, v in pairs(motif.files.font) do --loop through table keys
	if v ~= '' and motif.font_data[v] == nil then
		if motif.files.font_height[k] ~= nil then
			motif.font_data[v] = fontNew(v, motif.files.font_height[k])
		else
			motif.font_data[v] = fontNew(v)
		end
	end
	main.loadingRefresh()
end

local function f_facing(var)
	if var == -1 then
		return 'H'
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
	{s = 'stage_portrait_random_',        x = t.stage_pos[1] + t.stage_portrait_offset[1],         y = t.stage_pos[2] + t.stage_portrait_offset[2]},
	
}
for i = 1, #t_dir do
	--if t[t_dir[i].s .. 'offset'] == nil then t[t_dir[i].s .. 'offset'] = {0, 0} end
	--if t[t_dir[i].s .. 'scale'] == nil then t[t_dir[i].s .. 'scale'] = {1.0, 1.0} end
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
		t.continue_scale[2],
		'0',
		1,
		motif.defaultConnecting
	)
	animSetWindow(t.continue_anim_data, main.screenOverscan, 0, motif.info.localcoord[1], motif.info.localcoord[2])
end

if motif.vs_screen.p1_name_active_font == nil then
	motif.vs_screen.p1_name_active_font = {
		motif.vs_screen.p1_name_font[1],
		motif.vs_screen.p1_name_font[2],
		motif.vs_screen.p1_name_font[3],
		motif.vs_screen.p1_name_font[4],
		motif.vs_screen.p1_name_font[5],
		motif.vs_screen.p1_name_font[6],
		motif.vs_screen.p1_name_font[7],
		motif.vs_screen.p1_name_font[8]
	}
	motif.vs_screen.p1_name_active_font_scale = {motif.vs_screen.p1_name_font_scale[1], motif.vs_screen.p1_name_font_scale[2]}
end
if motif.vs_screen.p2_name_active_font == nil then
	motif.vs_screen.p2_name_active_font = {
		motif.vs_screen.p2_name_font[1],
		motif.vs_screen.p2_name_font[2],
		motif.vs_screen.p2_name_font[3],
		motif.vs_screen.p2_name_font[4],
		motif.vs_screen.p2_name_font[5],
		motif.vs_screen.p2_name_font[6],
		motif.vs_screen.p2_name_font[7],
		motif.vs_screen.p2_name_font[8]
	}
	motif.vs_screen.p2_name_active_font_scale = {motif.vs_screen.p2_name_font_scale[1], motif.vs_screen.p2_name_font_scale[2]}
end

main.f_printTable(motif, "debug/t_motif.txt")

return motif
