package event_room

import (
	"fmt"
	"sort"
	"math"
	"os/exec"
	"database/sql"
    _ "github.com/go-sql-driver/mysql"
	"encoding/json"
	"time"
	//"strings"
 	m "erps-go/msg"
	r "erps-go/room"
	e "erps-go/elo" 
)

var SCORE_INTERVAL int

type CreateRoomData struct {
	Id string
}

type CloseRoomData struct {
	Id string
}

type InviteRoomData struct {
	Room string
	Invite string
	From string
}

type JoinRoomData struct {
	Room string
	Join string
}

type UserLoginData struct {
	Id string
}

type UserNGHeroData struct {
	Id string
	Hero string
}

type UserLogoutData struct {
	Id string
}

type StartQueueData struct {
	Id string
	Action string
	Mode string
}

type CancelQueueData struct {
	Id string
	Action string
}

type PreGameData struct {
	Rid []([]uint32)
	Mode string
}

type PrestartData struct {
	Room string
	Id string
	Accept bool
}

type PrestartGetData struct {
	Room string
	Id string
}

type LeaveData struct {
	Room string
	Id string
}

type StartGameData struct {
	Game uint32
	Action string
}

type StartGameSendData struct {
	Game uint32
	Member []Herocell
}

type Herocell struct {
	Id string
	Team uint
	Name string
	Hero string
	Buff map[string]float32
	Tags []string
}

type GameOverData struct {
	Game uint32
	Win []string
	Lose []string
}

type GameCloseData struct {
	Game uint32
}

type StatusData struct {
	Id string
}

type ReconnectData struct {
	Id string
}

type ServerDead struct {
	ServerDead string
}

type GameInforData struct {
	Game uint32
	Users []UserInfoData
}

type UserInfoData struct {
	Id string
	Hero string
	Level uint16
	Equ []string
	Damage uint16
	Take_damage uint16
	Heal uint16
	Kill uint16
	Death uint16
	Assist uint16
	Gift UserGift
}

type UserGift struct {
	a uint16
	b uint16
	c uint16
	d uint16
	e uint16
}

type SqlLoginData struct {
	Id string
}

type SqlScoreData struct {
	Id string
	Score int16
	Mode string
}

type SqlGameInfoData struct {
	Game uint32
	Id string
	Hero string
	Level uint16
	Equ string
	Damage uint16
	Take_damage uint16
	Heal uint16
	Kill uint16
	Death uint16
	Assist uint16
	Gift UserGift
}

type QueueRoomData struct {
	User_name []string
	Rid uint32
	Gid uint32
	User_len uint16
	Avg_ng1v1 int16
	Avg_rk1v1 int16
	Avg_ng5v5 int16
	Avg_rk5v5 int16
	Mode string
	Ready uint
	Queue_cnt uint16
	Block []string
	Blacklist []string
}

type ReadyGroupData struct {
	User_name []string
	Gid uint32
	Rid []uint32
	User_len uint16
	Avg_ng1v1 int16
	Avg_rk1v1 int16
	Avg_ng5v5 int16
	Avg_rk5v5 int16
	Game_status uint16
	Queue_cnt uint16
	Block []string
	Blacklist []string
}

type ReadyGameData struct {
	User_name []string
	Gid []uint32
	Group []([]uint32)
	Team_len uint
	Block []string
}

type RemoveRoomData struct {
	Rid uint32
}


func SendGameList(game *r.FightGame, msgtx chan<-m.MqttMsg, conn *sql.DB) {
	res := StartGameSendData{};
	res.Game = game.Game_id
	for i, t := range game.Teams {
		ids := t.Get_users_id_hero()
		for _, j := range ids {
			h := Herocell {Id: j.Id, Team: uint(i+1), Name: j.Name, Hero: j.Hero}
			res.Member = append(res.Member, h)
		}
	}
	data, _ := json.Marshal(res)
	topic := fmt.Sprintf("game/%d/res/start_game", res.Game)
	msg := string(data)
	msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
}

func get_rid_by_id(id string, users map[string]*r.User) uint32 {
	if (users[id] != nil) {
		return users[id].Rid
	}
	return 0
}

func get_gid_by_id(id string, users map[string]*r.User) uint32 {
	if (users[id] != nil) {
		return users[id].Gid
	}
	return 0
}

func get_game_id_by_id(id string, users map[string]*r.User) uint32 {
	if (users[id] != nil) {
		return users[id].Game_id
	}
	return 0
}

func get_users(ids []string, users map[string]*r.User) []*r.User {
	res := []*r.User{}
	for _, i := range ids {
		res = append(res, users[i])
	}
	return res
}

func user_score(u *r.User, value int16, msgtx chan<-m.MqttMsg, sender chan<- m.SqlMsg, conn *sql.DB, mode string) {
	topic := fmt.Sprintf("member/%d/res/login", u.Id)
	msg := `{"msg":"ok"}`
	msgtx <- m.MqttMsg{Topic: topic, Msg: msg}

	data := SqlScoreData{}
	data.Id = u.Id;
	data.Mode = mode;
	if (mode == "ng1p2t") {
		data.Score = u.Ng1v1 + value
		d1, _ := json.Marshal(data)
        sender <- m.SqlMsg{Event: m.UpdateScore, Msg: string(d1)}
    
    } else if (mode == "ng5p2t") {
        data.Score = u.Ng5v5 + value
		d1, _ := json.Marshal(data)
        sender <- m.SqlMsg{Event: m.UpdateScore, Msg: string(d1)}
    
    } else if (mode == "rk1p2t") {
        data.Score = u.Rk1v1 + value
		d1, _ := json.Marshal(data)
        sender <- m.SqlMsg{Event: m.UpdateScore, Msg: string(d1)}
    
    } else if (mode == "rk5p2t") {
        data.Score = u.Rk5v5 + value
		d1, _ := json.Marshal(data)
        sender <- m.SqlMsg{Event: m.UpdateScore, Msg: string(d1)}
    
    }
}

func get_ng(team []*r.User, mode string) []int32 {
	res := []int32{}
	if (mode == "ng1p2t") {
		for _, u := range team {
			res = append(res, int32(u.Ng1v1))
		}
	} else if (mode == "ng5p2t") {
		for _, u := range team {
			res = append(res, int32(u.Ng5v5))
		}
	}
	return res
}

func get_rk(team []*r.User, mode string) []int32 {
	res := []int32{}
	if (mode == "rk1p2t") {
		for _, u := range team {
			res = append(res, int32(u.Rk1v1))
		}
	} else if (mode == "rk5p2t") {
		for _, u := range team {
			res = append(res, int32(u.Rk5v5))
		}
	}
	return res
}

func settlement_ng_score(win []*r.User, lose []*r.User, msgtx chan<-m.MqttMsg, sender chan<-m.SqlMsg, conn *sql.DB, mode string) {
	if (len(win) == 0 || len(lose) == 0) {
		return
	} 
	win_score := []int32{}
	lose_score := []int32{}
	if (mode == "ng1p2t" || mode == "ng5p2t") {
		win_score = get_ng(win, mode)
		lose_score = get_ng(lose, mode)
	} else {
		win_score = get_rk(win, mode)
		lose_score = get_rk(lose, mode)
	}
	elo := e.EloRank{K: 20.0}
	rw, rl := elo.Compute_elo_team(win_score, lose_score)
	for i, u := range win {
		user_score(u, int16(rw[i]-win_score[i]), msgtx, sender, conn, mode)
	}
	for i, u := range lose {
		user_score(u, int16(rl[i]-lose_score[i]), msgtx, sender, conn, mode)
	}
}

var ServerChan = make(chan m.ServerMsg)
var SqlChan = make(chan m.SqlMsg)


func HandleSqlRequest(conn *sql.DB) {
	
	update1000ms := time.NewTicker(1000*time.Millisecond)
	NewUsers := []string{}
	//len := 0
	UpdateInfo := []SqlGameInfoData{}
	//info_len := 0

	go func(){
		for {
			select {
				case <- update1000ms.C:
					if (len(NewUsers) > 0) {
						fmt.Println("Update")
						
						// insert new user into sql
						tx, _ := conn.Begin()
						for _, u := range NewUsers {
							tx.Exec("insert into user (userid, name, status) values(?,?,?)", u, "default name", "online")

						}
						tx.Commit()

						rows, _ := conn.Query("select id from user where userid=?", NewUsers[0])
						var id int
						for rows.Next() {
							rows.Scan(&id)
						}

						tx, _ = conn.Begin()
						for len, _ := range NewUsers {
							tx.Exec("insert into user_rk1v1 (id, score) values(?, 1000)", id+len)
							tx.Exec("insert into user_rk5v5 (id, score) values(?, 1000)", id+len)
							tx.Exec("insert into user_ng1v1 (id, score) values(?, 1000)", id+len)
							tx.Exec("insert into user_ng5v5 (id, score) values(?, 1000)", id+len)
						}
						tx.Commit()

						NewUsers = []string{}
					}
					if (len(UpdateInfo) > 0) {
						// inserat game info into sql
					}

				case val := <- SqlChan:
					if (val.Event == m.SqlLogin) {
						NewUsers = append(NewUsers, val.Msg)

					} else if (val.Event == m.UpdateScore) {

					} else if (val.Event == m.UpdateGameInfo) {

					}

			}
		}
	}()
	
}

func HandleQueueRequest(msgtx chan<-m.MqttMsg, mode string, team_size int16, match_size uint16) chan<-m.QueueMsg {
	QueueChan := make(chan m.QueueMsg)

	update1000ms := time.NewTicker(1000*time.Millisecond)
	go func() {
		QueueRoom := map[uint32]*QueueRoomData{}
		ReadyGroups := map[uint32]*ReadyGroupData{}

		group_id := uint32(0)
		for {
			select {
				case <- update1000ms.C:
					if (uint16(len(QueueRoom)) >= match_size) {
						g := new(ReadyGroupData)
						tq := []*QueueRoomData{}
						id := []uint32{}
						if (mode == "ng1p2t") {
							sort.Slice(tq, func(i, j int) bool { return tq[i].Avg_ng1v1 < tq[j].Avg_ng1v1 })
						} else if (mode == "rk1p2t") {
							sort.Slice(tq, func(i, j int) bool { return tq[i].Avg_rk1v1 < tq[j].Avg_rk1v1 })
						} else if (mode == "ng5p2t") {
							sort.Slice(tq, func(i, j int) bool { return tq[i].Avg_ng5v5 < tq[j].Avg_ng5v5 })
						} else if (mode == "rk5p2t") {
							sort.Slice(tq, func(i, j int) bool { return tq[i].Avg_rk5v5 < tq[j].Avg_rk5v5 })
						}
						for _, qr := range QueueRoom {
							block := false
							for _, user := range qr.User_name {
								if (block) {
									break
								}
								for _, id := range g.Blacklist {
									if (user == id) {
										block = true;
										break
									}
									if (block) {
										break
									}
								}
								for _, id := range g.Block {
									if (user == id) {
										block = true;
										break
									}
									if (block) {
										break
									}
								}
							}
							for _, user := range g.User_name {
								if (block) {
									break
								}
								for _, id := range qr.Blacklist {
									if (user == id) {
										block = true;
										break
									}
									if (block) {
										break
									}
								}
								for _, id := range qr.Block {
									if (user == id) {
										block = true;
										break
									}
									if (block) {
										break
									}
								}
							}
							if (block) {
								continue
							}
							group_score := int16(0)
							if (mode == "ng1p2t") {
								group_score = g.Avg_ng1v1
							} else if (mode == "rk1p2t") {
								group_score = g.Avg_rk1v1
							} else if (mode == "ng5p2t") {
								group_score = g.Avg_ng5v5
							} else if (mode == "rk5p2t") {
								group_score = g.Avg_rk5v5
							}

							room_score := int16(0)
							if (mode == "ng1p2t") {
								room_score = qr.Avg_ng1v1
							} else if (mode == "rk1p2t") {
								room_score = qr.Avg_rk1v1
							} else if (mode == "ng5p2t") {
								room_score = qr.Avg_ng5v5
							} else if (mode == "rk5p2t") {
								room_score = qr.Avg_rk5v5
							}

							if (g.User_len > 0 && g.User_len < uint16(team_size) && (uint16(group_score) + qr.Queue_cnt*uint16(SCORE_INTERVAL)) < uint16(room_score)) {
								for _, r := range g.Rid {
									id = append(id, r)
								}
								g = new(ReadyGroupData)
								g.Rid = append(g.Rid, qr.Rid)
								g.Block = append(g.Block, qr.Block...)
								g.Blacklist = append(g.Blacklist, qr.Blacklist...)
								g.User_name = append(g.User_name, qr.User_name...)

								score := int16(group_score * int16(g.User_len) + room_score * int16(qr.User_len)) / int16(g.User_len + qr.User_len)
								if (mode == "ng1p2t") {
									g.Avg_ng1v1 = score
								} else if (mode == "rk1p2t") {
									g.Avg_rk1v1 = score
								} else if (mode == "ng5p2t") {
									g.Avg_ng5v5 = score
								} else if (mode == "rk5p2t") {
									g.Avg_rk5v5 = score
								}
								g.User_len += qr.User_len
								qr.Ready = 1
								qr.Gid = group_id +1
								qr.Queue_cnt += 1
							}
							if (qr.Ready == 0 && int16(qr.User_len) + int16(g.User_len) <= team_size) {
								Difference := int(math.Abs(float64(room_score - group_score)))
								if (group_score == 0 || Difference <= SCORE_INTERVAL * int(qr.Queue_cnt)) {
									g.Rid = append(g.Rid, qr.Rid)
									g.Block = append(g.Block, qr.Block...)
									g.User_name = append(g.User_name, qr.User_name...)

									score := int16(group_score * int16(g.User_len) + room_score * int16(qr.User_len)) / int16(g.User_len + qr.User_len)
									if (mode == "ng1p2t") {
										g.Avg_ng1v1 = score
									} else if (mode == "rk1p2t") {
										g.Avg_rk1v1 = score
									} else if (mode == "ng5p2t") {
										g.Avg_ng5v5 = score
									} else if (mode == "rk5p2t") {
										g.Avg_rk5v5 = score
									}
									g.User_len += qr.User_len
									qr.Ready = 1
									qr.Gid = group_id + 1
								} else {
									qr.Queue_cnt += 1
								}
							} 
							if (g.User_len == uint16(team_size)) {
								group_id += 1
								g.Gid = group_id
								g.Queue_cnt = 1
								ReadyGroups[group_id] = g
								g = new(ReadyGroupData)
							}
						}
					}
					if (uint16(len(ReadyGroups)) >= match_size) {
						fg := new(ReadyGameData)
						total_score := int16(0)
						rm_ids := []uint32{}

						for id, rg := range ReadyGroups {
							block := false
							for _, user := range rg.User_name {
								if (block) {
									break
								}
								for _, id := range fg.Block {
									if (user == id) {
										block = true;
										break
									}
									if (block) {
										break
									}
								}
							}
							for _, user := range fg.User_name {
								if (block) {
									break
								}
								for _, id := range rg.Block {
									if (user == id) {
										block = true;
										break
									}
									if (block) {
										break
									}
								}
							}
							if (block) {
								continue
							}
							group_score := int16(0)
							if (mode == "ng1p2t") {
								group_score = rg.Avg_ng1v1
							} else if (mode == "rk1p2t") {
								group_score = rg.Avg_rk1v1
							} else if (mode == "ng5p2t") {
								group_score = rg.Avg_ng5v5
							} else if (mode == "rk5p2t") {
								group_score = rg.Avg_rk5v5
							}

							if (rg.Game_status == 0 && uint16(fg.Team_len) < match_size) {
								if (total_score == 0) {
									total_score += group_score
									fg.Group = append(fg.Group, rg.Rid)
									fg.Gid = append(fg.Gid, id)
									fg.Team_len += 1
									fg.Block = append(fg.Block, rg.Block...)
									continue
								}

								Difference := int16(0)
								if (fg.Team_len > 0) {
									Difference = int16(math.Abs(float64(group_score - total_score/int16(fg.Team_len))))
								}
								if (Difference <= int16(SCORE_INTERVAL) * int16(rg.Queue_cnt)) {
									total_score += group_score
									fg.Group = append(fg.Group, rg.Rid)
									fg.Block = append(fg.Block, rg.Block...)
									fg.User_name = append(fg.User_name, rg.User_name...)
									fg.Team_len += 1
									fg.Gid = append(fg.Gid, id)
								} else {
									rg.Queue_cnt += 1
								}
							}
							if (uint16(fg.Team_len) == match_size) {
								data := PreGameData {
									Rid: fg.Group,
									Mode: mode,
								}
								d, _ := json.Marshal(data)
								msg := string(d)
								ServerChan <- m.ServerMsg{Event: m.UpdateGame, Id: "0", Msg: msg}
								for _, id := range fg.Gid {
									rm_ids = append(rm_ids, id)
								}
								fg = new(ReadyGameData)
							}
						}
						for _, id := range rm_ids {
							for _, rid := range ReadyGroups[id].Rid {
								delete(QueueRoom, rid)
							}
							delete(ReadyGroups, id)
						}
					}


				case val := <- QueueChan:
					if (val.Event == m.UpdateRoom) {
						fmt.Println("Update Room")
						p := &QueueRoomData{}
						json.Unmarshal([]byte(val.Msg), &p)
						QueueRoom[p.Rid] = p
					} else if (val.Event == m.RemoveRoom) {
						fmt.Println("Update Room")
						p := &RemoveRoomData{}
						json.Unmarshal([]byte(val.Msg), &p)
						room, ok_r := QueueRoom[p.Rid]
						if (ok_r) {
							rg, ok_rg := ReadyGroups[room.Gid]
							if (ok_rg) {
								for _, rid := range rg.Rid {
									if (rid == p.Rid) {
										continue
									}
									r, ok_rr := QueueRoom[rid]
									if (ok_rr) {
										r.Gid = 0
										r.Ready = 0
									}
								}
							}
							delete(ReadyGroups, room.Gid)
						}
						delete(QueueRoom, p.Rid)
					}
			}
		}
	}()
	return QueueChan
}

func Init(msgtx chan<-m.MqttMsg, conn *sql.DB) {
	update5000ms := time.NewTicker(5000*time.Millisecond)
	update200ms := time.NewTicker(200*time.Millisecond)
	update1000ms := time.NewTicker(1000*time.Millisecond)
	QueueSender := map[string]chan<-m.QueueMsg{}

	HandleSqlRequest(conn)
	QueueSender["ng1p2t"] = HandleQueueRequest(msgtx, "ng1p2t", 1, 2)
	QueueSender["ng5p2t"] = HandleQueueRequest(msgtx, "ng5p2t", 5, 2)
	QueueSender["rk1p2t"] = HandleQueueRequest(msgtx, "rk1p2t", 1, 2)
	QueueSender["rk5p2t"] = HandleQueueRequest(msgtx, "rk5p2t", 5, 2)
	SCORE_INTERVAL = 100

	go func() {
		// User
		TotalUsers := map[string]*r.User{}
		// Room
		TotalRoom := map[uint32]*r.RoomData{}
		ReadyGroups := map[uint32]*r.FightGroup{}
		PreStartGroups := map[uint32]*r.FightGame{}
		GamingGroup := map[uint32]*r.FightGame{}
		room_id := uint32(0)
		game_id := uint32(0)
		group_id := uint32(0)
		game_port := uint16(7777)
		// load user from sql
		rows, _ := conn.Query(`select userid, a.score as ng1v1, b.score as rk1v1, d.score as ng5v5, e.score as rk5v5, name from user as c 
								join user_ng1v1 as a on a.id=c.id 
								join user_rk1v1 as b on b.id=c.id
								join user_ng5v5 as d on d.id=c.id
								join user_rk5v5 as e on e.id=c.id `)
		defer rows.Close()
		for rows.Next() {
			var id string
			var hero string
			var ng1v1 int16
			var ng5v5 int16
			var rk1v1 int16
			var rk5v5 int16
			if err := rows.Scan(&id, &ng1v1, &rk1v1, &ng5v5, &rk5v5, &hero); err != nil {
				fmt.Println(err)
			}
			TotalUsers[id] = &r.User{
				Id: id,
				Name: hero,
				Hero: "",
				Ng1v1: ng1v1,
				Ng5v5: ng5v5,
				Rk1v1: rk1v1,
				Rk5v5: rk5v5,
				Rid: 0,
				Gid: 0,
				Game_id: 0,
				Online: false,
				Start_prestart: false,
				Prestart_get: false,
				Recent_users: [][]string{},
				Blacklist: []string{},
			}
			//fmt.Println("Id: %s, hero: %s, ng1v1: %d, ng5v5: %d, rk1v1: %d, rk5v5: %d", id, hero, ng1v1, ng5v5, rk1v1, rk5v5)
		}

		rows, _ = conn.Query(`select * from user_blacklist `)
		defer rows.Close()
		for rows.Next() {
			var id string
			var black string
			if err := rows.Scan(&id, &black); err != nil {
				fmt.Println(err)
			}
			user, ok_u := TotalUsers[id]
			if (ok_u) {
				user.Blacklist = append(user.Blacklist, black)
			}
		}
		fmt.Println("ok")

		for {
			select {
				case <- update200ms.C:
					// prestart check
					rm_ids := []uint32{}
					start_cnt := 0
					for id, group := range PreStartGroups {
						res := group.Check_prestart()
						switch (res) {
						case r.Ready:
							start_cnt += 1
							rm_ids = append(rm_ids, id)
							game_port += 1;
							if (game_port > 65500) {
								game_port = 7777
							}
							group.Ready()
							group.Update_names()
							group.Game_port = game_port

							delete(GamingGroup, group.Game_id)
							GamingGroup[group.Game_id] = group

							app := "/home/damody/LinuxNoEditor/CF1/Binaries/Linux/CF1Server"
							arg0 := fmt.Sprintf("-Port=%s", game_port)
							arg1 := fmt.Sprintf("-gameid %s", group.Game_id)
							arg2 := `-NOSTEAM`
							cmd := exec.Command(app, arg0, arg1, arg2)
							stdout, err := cmd.Output()
							if err != nil {
								fmt.Println(err.Error())
								//return
							}
						
							fmt.Print(string(stdout))

						case r.Cancel:
							group.Update_names()
							group.Clear_queue()
							group.Game_status = 0
							rm_rid := []uint32{}
							users := []string{}
							block := []string{}
							blacklist := []string{}
							for _, t := range group.Teams {
								for _, c := range t.Checks {
									if (c.Check < 0) {
										user, ok_u := TotalUsers[c.Id]
										if (ok_u) {
											rm_rid = append(rm_rid, user.Rid)
										}
									}
								}
								for _, room := range t.Rooms {
									for _, user := range room.Users {
										users = append(users, user.Id)

										for _, ids := range user.Recent_users {
											block = append(block, ids...)
										}
										blacklist = append(blacklist, user.Blacklist...)
									}
								}
							}
							for _, room := range group.Room_names {
								topic := fmt.Sprintf("room/%s/res/prestart", room)
								msg := `{"msg":"stop queue"}`
								msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
							}
							for _, team := range group.Teams {
								for _, room := range team.Rooms {
									contains := false
									for _, rid := range rm_rid {
										if (rid == room.Rid) {
											contains = true
											break
										}
									}
									if (!contains) {
										t1, ok_t := QueueSender[group.Mode]
										if (ok_t) {
											data := QueueRoomData{
												User_name: users,
												Rid: room.Rid,
												Gid: 0,
												User_len: uint16(len(room.Users)),
												Avg_ng1v1: room.Avg_ng1v1,
												Avg_rk1v1: room.Avg_rk1v1,
												Avg_ng5v5: room.Avg_ng5v5,
												Avg_rk5v5: room.Avg_rk5v5,
												Mode: group.Mode,
												Ready: 0,
												Queue_cnt: 1,
												Block: block,
												Blacklist: blacklist,
											}
											d, _ := json.Marshal(data)
											msg := string(d)
											t1 <- m.QueueMsg{Event: m.UpdateRoom, Msg: msg}
										}
									}

								}
							}
							rm_ids = append(rm_ids, id)

						case r.Wait:
							
						}
					}
					for _, id := range rm_ids {
						delete(PreStartGroups, id)
					}

				case <- update1000ms.C:
					// heartbeat
				
				case <- update5000ms.C:
					// re-send prestart

				case val := <- ServerChan:
					if (val.Event == m.Login) {
						p := &UserLoginData{}
						json.Unmarshal([]byte(val.Msg), &p)
						fmt.Println("Login: ", p.Id)
						user, ok := TotalUsers[p.Id]
						// if Totaluser contains the ID, respond "ok"
						if (ok) {
							user.Online = true
							topic := fmt.Sprintf("member/%s/res/login", p.Id)
							msg := `{"msg":"ok"}`
							msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
							
							topic = fmt.Sprintf("member/%s/res/score", p.Id)
							msg = fmt.Sprintf(`{"ng1p2t":%d, "ng5p2t":%d, "rk1p2t":%d, "rk5p2t":%d}`, user.Ng1v1, user.Ng5v5, user.Rk1v1, user.Rk5v5)
							msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
							
						} else {
							TotalUsers[p.Id] = &r.User{
								Id: p.Id,
								Name: "default name",
								Hero: "",
								Ng1v1: 1000,
								Ng5v5: 1000,
								Rk1v1: 1000,
								Rk5v5: 1000,
								Rid: 0,
								Gid: 0,
								Game_id: 0,
								Online: false,
								Start_prestart: false,
								Prestart_get: false,
							}
							topic := fmt.Sprintf("member/%s/res/login", p.Id)
							msg := `{"msg":"ok"}`
							msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
							
							topic = fmt.Sprintf("member/%s/res/score", p.Id)
							msg = `{"ng1p2t":1000, "ng5p2t":1000, "rk1p2t":1000, "rk5p2t":1000}`
							msgtx <- m.MqttMsg{Topic: topic, Msg: msg}

							SqlChan <- m.SqlMsg{Event: m.SqlLogin, Msg: p.Id}
							
						}
						
					} else if (val.Event == m.Logout) {
						success := false
						p := &UserLogoutData{}
						json.Unmarshal([]byte(val.Msg), &p)
						fmt.Println("Logout: ", p.Id)
						user, ok := TotalUsers[p.Id]
						if (ok) {
							is_null := false
							user.Online = false
							if (user.Game_id == 0) {
								gid := user.Gid
								rid := user.Rid
								r, ok_r := TotalRoom[rid]

								if (ok_r) {
									master := r.Master
									r.Rm_user(p.Id)
									if (len(r.Users) > 0) {
										r.Publish_update(msgtx, master)
									} else {
										is_null = true
									}
									topic := fmt.Sprintf("room/%s/res/leave", p.Id)
									msg := `{"msg":"ok"}`
									msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
								}

								if (gid != 0) {
									g, ok_g := ReadyGroups[gid]
									if (ok_g) {
										g.User_cancel(p.Id)
										delete(ReadyGroups, gid)
										r, ok_r := TotalRoom[rid]
										if (ok_r) {
											// Send QueueRequest
											tx, ok := QueueSender[r.Mode]
											if (ok) {
												data := RemoveRoomData{Rid: rid}
												d, _ := json.Marshal(data)
												msg := string(d)
												tx <- m.QueueMsg{Event: m.RemoveRoom, Msg: msg}
											}

											topic := fmt.Sprintf("room/%s/res/cancel_queue", r.Master)
											msg := `{"msg":"ok"}`
											msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
										}
									}
								}
								success = true
							} else {
								success = false
							}
							if (is_null) {
								delete(TotalRoom, user.Rid)
							}
						}
						if (success) {
							topic := fmt.Sprintf("member/%s/res/logout", p.Id)
							msg := `{"msg":"ok"}`
							msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
						
						} else {
							topic := fmt.Sprintf("member/%s/res/logout", p.Id)
							msg := `{"msg":"fail"}`
							msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
						
						}

					} else if (val.Event == m.Create) {
						p := &CreateRoomData{}
						json.Unmarshal([]byte(val.Msg), &p)
						success := false
						_, ok := TotalRoom[get_rid_by_id(p.Id, TotalUsers)]
						if (!ok) {
							u , ok_u := TotalUsers[p.Id]
							if (ok_u) {
								room_id += 1
								TotalRoom[room_id] = &r.RoomData{
									Rid: room_id,
									Users: []*r.User{},
									Master: p.Id,
									Last_master: "",
									Mode: "",
									Avg_ng1v1: 0,
									Avg_ng5v5: 0,
									Avg_rk1v1: 0,
									Avg_rk5v5: 0,
									Ready: 0,
									Queue_cnt: 1,
								}
								TotalRoom[room_id].Add_user(u)
								TotalRoom[room_id].Publish_update(msgtx, p.Id)
								success = true
							}
						}
						if (success) {
							topic := fmt.Sprintf("room/%s/res/create", p.Id)
							msg := `{"msg":"ok"}`
							msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
						} else {
							topic := fmt.Sprintf("room/%s/res/create", p.Id)
							msg := `{"msg":"fail"}`
							msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
						}
					} else if (val.Event == m.Close) {	
					} else if (val.Event == m.ChooseNGHero) {
						p := &UserNGHeroData{}
						json.Unmarshal([]byte(val.Msg), &p)
						user, ok := TotalUsers[p.Id]
						if (ok) {
							user.Hero = p.Hero
							topic := fmt.Sprintf("member/%s/res/choose_hero", user.Id)
							msg := fmt.Sprintf(`{"id":"%s", "hero":"%s"}`, user.Id, user.Hero)
							msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
						}
					} else if (val.Event == m.Invite) {
						p := &InviteRoomData{}
						json.Unmarshal([]byte(val.Msg), &p)
						_, ok_u := TotalUsers[p.From]
						if (ok_u) {
							topic := fmt.Sprintf("room/%s/res/invite", p.Invite)
							msg := fmt.Sprintf(`{"room":"%s", "from":"%s"}`, p.Room, p.From)
							msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
						}
					} else if (val.Event == m.Join) {
						p := &JoinRoomData{}
						json.Unmarshal([]byte(val.Msg), &p)
						user, ok_u := TotalUsers[p.Room]
						user2, ok_j := TotalUsers[p.Join]
						sendok := false
						if (ok_u) {
							if (ok_j) {
								room, ok_r := TotalRoom[user.Rid]
								if (ok_r) {
									if (room.Ready == 0 && len(room.Users) < 5) {
										room.Add_user(user2)
										master := room.Master
										room.Publish_update(msgtx, master)
										room.Publish_update(msgtx, p.Join)
										topic := fmt.Sprintf("room/%s/res/join", p.Join)
										msg := fmt.Sprintf(`{"room":"%s", "msg":"ok"}`, room.Master)
										msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
										sendok = true
									}
								}
							}
						}
						if (!sendok) {
							topic := fmt.Sprintf("room/%s/res/join", p.Join)
							msg := fmt.Sprintf(`{"room":"%s", "msg":"fail"}`, p.Room)
							msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
						
						}

					} else if (val.Event == m.StartQueue) {
						p := &StartQueueData{}
						json.Unmarshal([]byte(val.Msg), &p)
						success := false
						hasRoom := false
						user, ok_u := TotalUsers[p.Id]
						rid := uint32(0)
						if (ok_u) {
							if (user.Rid != 0) {
								hasRoom = true
								rid = user.Rid
							}
						}
						if (hasRoom) {
							room, ok_r := TotalRoom[rid]
							if (ok_r) {
								if ((p.Mode == "ng1p2t" || p.Mode == "rkip2t") && len(room.Users) > 1) {
									topic := fmt.Sprintf("room/%s/res/start_queue", room.Master)
									msg := `{"msg":"fail"}`
									msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
								} else {
									users := []string{}
									block := []string{}
									blacklist := []string{}
									for _, user := range room.Users {
										users = append(users, user.Id)

										for _, ids := range user.Recent_users {
											block = append(block, ids...)
										}
										blacklist = append(blacklist, user.Blacklist...)
									}
									room.Mode = p.Mode
									room.Update_avg()
									t1, ok_t := QueueSender[p.Mode]
									if (ok_t) {
										data := QueueRoomData{
											User_name: users,
											Rid: room.Rid,
											Gid: 0,
											User_len: uint16(len(room.Users)),
											Avg_ng1v1: room.Avg_ng1v1,
											Avg_rk1v1: room.Avg_rk1v1,
											Avg_ng5v5: room.Avg_ng5v5,
											Avg_rk5v5: room.Avg_rk5v5,
											Mode: p.Mode,
											Ready: 0,
											Queue_cnt: 1,
											Block: block,
											Blacklist: blacklist,
										}
										d, _ := json.Marshal(data)
										msg := string(d)
										t1 <- m.QueueMsg{Event: m.UpdateRoom, Msg: msg}
									}
									success = true
									if (success) {
										topic := fmt.Sprintf("room/%s/res/start_queue", room.Master)
										msg := `{"msg":"ok"}`
										msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
									} else {
										topic := fmt.Sprintf("room/%s/res/start_queue", room.Master)
										msg := `{"msg":"fail"}`
										msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
									}
								}
							}
						}
					} else if (val.Event == m.CancelQueue) {
						p := &CancelQueueData{}
						json.Unmarshal([]byte(val.Msg), &p)
						success := false
						user, ok_u := TotalUsers[p.Id]
						if (ok_u) {
							g, ok_g := ReadyGroups[user.Gid]
							if (ok_g) {
								for _, room := range g.Rooms {
									room.Ready = 0
								}
							}
							//rg, ok_rg := ReadyGroups[user.Gid]
							room, ok_r := TotalRoom[user.Rid]
							if (ok_r) {
								t1, ok_t := QueueSender[room.Mode]
								if (ok_t) {
									data := RemoveRoomData{
										Rid: user.Rid,
									}
									d, _ := json.Marshal(data)
									msg := string(d)
									t1 <- m.QueueMsg{Event: m.RemoveRoom, Msg: msg}
								}
								success = true;
								if (success) {
									topic := fmt.Sprintf("room/%s/res/cancel_queue", room.Master)
									msg := `{"msg":"ok"}`
									msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
								} else {
									topic := fmt.Sprintf("room/%s/res/cancel_queue", room.Master)
									msg := `{"msg":"fail"}`
									msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
								}
							}
						}
					} else if (val.Event == m.UpdateGame) {
						fmt.Println("Update Room")
						p := &PreGameData{}
						json.Unmarshal([]byte(val.Msg), &p)
						fmt.Println("Update Game")
						fg := new(r.FightGame)
						for _, r1 := range p.Rid {
							g := new(r.FightGroup)
							for _, rid := range r1 {
								room, ok_r := TotalRoom[rid]
								if (ok_r) {
									g.Add_room(room)
								}
							}
							g.Prestart()
							group_id += 1
							g.Set_group_id(group_id)
							g.Game_status = 1
							g.Mode = p.Mode
							ReadyGroups[group_id] = g
							rg, ok_rg := ReadyGroups[group_id]
							if (ok_rg) {
								fg.Teams = append(fg.Teams, rg)
							}
						}

						fg.Update_names()
						for _, r := range fg.Room_names {
							topic := fmt.Sprintf("room/%s/res/prestart", r)
							msg := `{"msg":"prestart"}`
							msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
						}
						fg.Mode = p.Mode
						game_id += 1
						fg.Set_game_id(game_id)
						PreStartGroups[game_id] = fg

					} else if (val.Event == m.PreStart) {
						p := &PrestartData{}
						json.Unmarshal([]byte(val.Msg), &p)
						fmt.Println("Prestart")
						fmt.Println(p.Accept)
						user, ok_u := TotalUsers[p.Room]
						if (ok_u) {
							gid := user.Gid
							if (user.Prestart_get == true) {
								if (gid != 0) {
									g, ok_g := ReadyGroups[gid]
									if (ok_g) {
										if (p.Accept == true) {
											fmt.Println("Prestart")
						
											g.User_ready(p.Id)
											topic := fmt.Sprintf("room/%s/res/start_get", user.Id)
											msg := `{"msg":"start"}`
											msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
										} else {
											g.User_cancel(p.Id)
											users := []string{}
											block := []string{}
											blacklist := []string{}
											for _, room := range g.Rooms{
												for _, user := range room.Users {
													users = append(users, user.Id)

													for _, ids := range user.Recent_users {
														block = append(block, ids...)
													}
													blacklist = append(blacklist, user.Blacklist...)
												}
											}
											for _, room := range g.Rooms {
												if (room.Rid != user.Rid) {
													data := QueueRoomData{
														User_name: users,
														Rid: room.Rid,
														Gid: 0,
														User_len: uint16(len(room.Users)),
														Avg_ng1v1: room.Avg_ng1v1,
														Avg_rk1v1: room.Avg_rk1v1,
														Avg_ng5v5: room.Avg_ng5v5,
														Avg_rk5v5: room.Avg_rk5v5,
														Mode: g.Mode,
														Ready: 0,
														Queue_cnt: 1,
														Block: block,
														Blacklist: blacklist,
													}
													t1, ok_t := QueueSender[g.Mode]
													if (ok_t) {
														d, _ := json.Marshal(data)
														msg := string(d)
														t1 <- m.QueueMsg{Event: m.UpdateRoom, Msg: msg}
													}
												}
											}
											delete(ReadyGroups, gid)
											room, ok_r := TotalRoom[user.Rid]
											if (ok_r) {
												topic := fmt.Sprintf("room/%s/res/cancel_queue", room.Master)
												msg := `{"msg":"ok"}`
												msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
											}
										}
									}
								}
							}
						}
					} else if (val.Event == m.PreStartGet) {
						fmt.Println("Prestart Get")
						p := &PrestartGetData{}
						json.Unmarshal([]byte(val.Msg), &p)
						user, ok_u := TotalUsers[p.Id]
						if (ok_u) {
							user.Prestart_get = true
						}
					} else if (val.Event == m.Leave) {
						p := &LeaveData{}
						json.Unmarshal([]byte(val.Msg), &p)
						user, ok_u := TotalUsers[p.Id]
						if (ok_u) {
							room, ok_r := TotalRoom[user.Rid]
							is_null := false
							if (ok_r) {
								
								//fmt.Println("in")
								master := room.Master
								room.Rm_user(p.Id)
								if (len(room.Users) > 0) {
									room.Publish_update(msgtx, master)
								} else {
									is_null = true
								}
								topic := fmt.Sprintf("room/%s/res/leave", p.Id)
								msg := fmt.Sprintf(`{"msg":"ok", "id":"%s"}`, p.Id)
								msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
								
								if (is_null) {
									if (room.Mode != "") {
										// send queuehandle
										t1, ok_t := QueueSender[room.Mode]
										if (ok_t) {
											data := RemoveRoomData{Rid: user.Rid}
											d, _ := json.Marshal(data)
											msg := string(d)
											t1 <- m.QueueMsg{Event: m.RemoveRoom, Msg: msg}
										}
									}
								}
							}
							if (is_null) {
								delete(TotalRoom, user.Rid)
							}
							user.Rid = 0
						}
					} else if (val.Event == m.StartGame) {
						
					} else if (val.Event == m.GameOver) {
						
					} else if (val.Event == m.GameInfo) {
						
					} else if (val.Event == m.GameClose) {
						
					} else if (val.Event == m.Status) {
						p := &StatusData{}
						json.Unmarshal([]byte(val.Msg), &p)
						user, ok := TotalUsers[p.Id]
						if (ok) {
							if (user.Game_id != 0) {
								topic := fmt.Sprintf("member/%s/res/status", p.Id)
								msg := `{"msg":"gaming"}`
								msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
							} else {
								topic := fmt.Sprintf("member/%s/res/status", p.Id)
								msg := `{"msg":"normal"}`
								msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
							}
						} else {
							topic := fmt.Sprintf("member/%s/res/status", p.Id)
								msg := `{"msg":"id not found"}`
								msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
						}
						
					} else if (val.Event == m.Reconnect) {
						p := &ReconnectData{}
						json.Unmarshal([]byte(val.Msg), &p)
						user, ok_u := TotalUsers[p.Id]
						if (ok_u) {
							g, ok_g := GamingGroup[user.Game_id]
							if (ok_g) {
								topic := fmt.Sprintf("member/%s/res/reconnect", p.Id)
								msg := fmt.Sprintf(`{"server":"114.32.129.195:%s"}`, g.Game_port)
								msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
							}
						}
					} else if (val.Event == m.MainServerDead) {
						
					}

			}

		}
	}()
}