--
alter table clubs add club_coin BIGINT default 0 not null after diamond;
alter table clubs add notice VARCHAR(2000) NOT NULL DEFAULT '' after club_param;
alter table clubs add setting_param VARCHAR(200) NOT NULL DEFAULT '' after notice;

alter table club_members add club_coin BIGINT default 0 not null after status;

alter table club_journals add status INT default 110 not null after `foreign`;

alter table rooms add sub_room_type INT default 0 not null after cost_type;
alter table rooms add setting_param VARCHAR(200) NOT NULL DEFAULT '' after sub_room_type;
alter table rooms add start_max_number INT default 0 not null after sub_room_type;

alter table users add register_channel VARCHAR(200) NOT NULL DEFAULT '' after last_login_ip;



alter table thirteens modify column user_cards VARCHAR(1200);
alter table thirteens modify column user_submit_cards VARCHAR(1200);
alter table thirteens modify column game_results VARCHAR(3000);