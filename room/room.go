package room

import (
	//"fmt"
	"encoding/json"
	//"math/rand"
	//"time"
	m "erps-go/msg"
)

type User struct {
	Id string
	Name string
	Hero string
	Ng1v1 int16
	Ng5v5 int16
	Rk1v1 int16
	Rk5v5 int16
	Rid uint32
	Gid uint32
	Game_id uint32
	Online bool
	Start_prestart bool
	Prestart_get bool
	Recent_users []([]string)
	Blacklist []string
}

type RoomData struct {
	Rid uint32
	Users []*User
	Master string
	Last_master string
	Mode string
	Avg_ng1v1 int16
	Avg_ng5v5 int16
	Avg_rk1v1 int16
	Avg_rk5v5 int16
	Ready int
	Queue_cnt int16
}

func (r *RoomData) Update_avg() {
	sum_ng1v1 := int16(0)
	sum_rk1v1 := int16(0)
	sum_ng5v5 := int16(0)
	sum_rk5v5 := int16(0)
	
	for _, u := range r.Users {
		sum_ng1v1 += u.Ng1v1
		sum_rk1v1 += u.Rk1v1
		sum_ng5v5 += u.Ng5v5
		sum_rk5v5 += u.Rk5v5
	}
	if (len(r.Users) > 0) {
		r.Avg_ng1v1 = int16(sum_ng1v1/int16(len(r.Users)))
		r.Avg_ng5v5 = int16(sum_ng5v5/int16(len(r.Users)))
		r.Avg_rk1v1 = int16(sum_rk1v1/int16(len(r.Users)))
		r.Avg_rk5v5 = int16(sum_rk5v5/int16(len(r.Users)))
	}
}

func (r *RoomData) Add_user(user *User) {
	user.Rid = r.Rid
	r.Users = append(r.Users, user)
	r.Update_avg()
}

func (r *RoomData) Leave_room() {
	for _, u := range r.Users {
		u.Rid = 0
		u.Gid = 0
		u.Game_id = 0
	}
	r.Ready = 0
}

func (r *RoomData) User_prestart() {
	for _, u := range r.Users {
		u.Start_prestart = true
		u.Prestart_get = false
	}
}

func (r *RoomData) Check_prestart_get() bool {
	res := false
	for _, u := range r.Users {
		if (u.Prestart_get) {
			res = true
		} else {
			res = false
			break;
		}
	}
	return res
}

func (r *RoomData) Clear_queue() {
	for _, u := range r.Users {
		u.Gid = 0
		u.Game_id = 0;
	}
	r.Ready = 0
}

func (r *RoomData) Publish_update(msgtx chan<-m.MqttMsg, rid string) {
	type TeamCell struct{
		Room string
		Team []string
	}
	t := TeamCell {Room: r.Master}

	for _, u := range r.Users {
		t.Team = append(t.Team, u.Id)
	}
	topic := "room/" + rid + "/res/update"
	d, _ := json.Marshal(t)
	msg := string(d)
	msgtx <- m.MqttMsg{Topic: topic, Msg: msg}
}

func (r *RoomData) Rm_user(Id string) {
	i := 0
	for {
		if ( i == len(r.Users)) {
			break
		}
		id2 := r.Users[i].Id
		if id2 == Id {
			r.Users = append(r.Users[:i], r.Users[i+1:]...)
			
		} else {
			i += 1
		}
	}
	if (r.Master == Id && len(r.Users) > 0) {
		r.Last_master = r.Master
		r.Master = r.Users[0].Id
	}
	r.Update_avg()
}

type FightCheck struct {
	Id string
	Check int
}

type FightGroup struct {
	Rooms []*RoomData
	User_count int16
	Avg_ng1v1 int16
	Avg_ng5v5 int16
	Avg_rk1v1 int16
	Avg_rk5v5 int16
	Mode string
	Checks []*FightCheck
	Rids []uint32
	Game_status uint16
	Queue_cnt int16
}

func (fg *FightGroup) User_ready(id string) bool {
	for _, c := range fg.Checks {
		if c.Id == id {
			c.Check = 1
			fg.Check_prestart()
			return true
		}
	}
	return false
}

type Users_Id_Hero struct {
	Id string
	Name string
	Hero string
}

func (fg *FightGroup) Get_users_id_hero() []*Users_Id_Hero {
	Uih := []*Users_Id_Hero{}
	for _, r := range fg.Rooms {
		for _, u := range r.Users {
			data := new(Users_Id_Hero)
			data.Id = u.Id
			data.Name = u.Name
			data.Hero = u.Hero
			Uih = append(Uih, data);
		}
	}
	return Uih
}

func (fg *FightGroup) User_cancel(id string) bool {
	for _, c := range fg.Checks {
		if c.Id == id {
			c.Check = -1
			return true
		}
	}
	return false
}

func (fg *FightGroup) Leave_room() bool {
	for _, r := range fg.Rooms {
		r.Leave_room()
	}
	return false
}

func (fg *FightGroup) Clear_queue() bool {
	for _, r := range fg.Rooms {
		r.Clear_queue()
	}
	return false
}

func (fg *FightGroup) Check_has_room(id string) bool {
	for _, r := range fg.Rooms {
		if (r.Master == id) {
			return true
		}
	}
	return false
}

func (fg *FightGroup) Update_avg() {
	sum_ng1v1 := int32(0)
	sum_rk1v1 := int32(0)
	sum_ng5v5 := int32(0)
	sum_rk5v5 := int32(0)
	fg.User_count = 0
	for _, r := range fg.Rooms {
		sum_ng1v1 += int32(r.Avg_ng1v1 * int16(len(r.Users)))
		sum_rk1v1 += int32(r.Avg_rk1v1 * int16(len(r.Users)))
		sum_ng5v5 += int32(r.Avg_ng5v5 * int16(len(r.Users)))
		sum_rk5v5 += int32(r.Avg_rk5v5 * int16(len(r.Users)))
	}
	if (fg.User_count > 0) {
		fg.Avg_ng1v1 = int16(sum_ng1v1/int32(fg.User_count))
		fg.Avg_ng5v5 = int16(sum_ng5v5/int32(fg.User_count))
		fg.Avg_rk1v1 = int16(sum_rk1v1/int32(fg.User_count))
		fg.Avg_rk5v5 = int16(sum_rk5v5/int32(fg.User_count))
	}
}

func (fg *FightGroup) Add_room(room *RoomData) {
	fg.Rooms = append(fg.Rooms, room)
	fg.Rids = append(fg.Rids, room.Rid)
	fg.Update_avg()
}

func (fg *FightGroup) Rm_room_by_master(id string) {
	new_rids := []uint32{}
	index := -1
	for idx, r := range fg.Rooms {
		if (r.Master == id) {
			for _, rid := range fg.Rids {
				if (rid != r.Rid) {
					new_rids = append(new_rids, rid)
				}
			}
			index = idx
			break
		}
	}
	if (index != -1) {
		copy(fg.Rooms[index:], fg.Rooms[index+1:]) 
		fg.Rooms[len(fg.Rooms)-1] = &RoomData{}     
		fg.Rooms = fg.Rooms[:len(fg.Rooms)-1]
	}     
	fg.Rids = new_rids
	fg.Update_avg()
}

func (fg *FightGroup) Rm_room_by_rid(id uint32) {
	//i := 0
	new_rids := []uint32{}
	new_rooms := []*RoomData{}
	for _, rid := range fg.Rids {
		if (rid != id) {
			new_rids = append(new_rids, rid)
		}
	}
	for _, r := range fg.Rooms {
		if (r.Rid != id) {
			new_rooms = append(new_rooms, r)
		}
	}
	fg.Rids = new_rids
	fg.Rooms = new_rooms
} 
func (fg *FightGroup) Ready() {
	fg.Checks = []*FightCheck{}
	for _, r := range fg.Rooms {
		r.Ready = 3
	}
}

func (fg *FightGroup) Set_group_id(gid uint32) {
	for _, r := range fg.Rooms {
		r.Ready = 1
		for _, u := range r.Users {
			u.Gid = gid
		}
	}
}

func (fg *FightGroup) Prestart() {
	fg.Checks = []*FightCheck{}
	for _, r := range fg.Rooms {
		r.Ready = 1
		r.User_prestart()
		for _, u := range r.Users {
			fg.Checks = append(fg.Checks, &FightCheck{Id: u.Id, Check: 0})
		}
	}
}

func (fg *FightGroup)  Check_prestart() int {
	res := Ready
	for _, c := range fg.Checks {
		if (c.Check < 0) {
			return Cancel
		} else if (c.Check != 1) {
			res = Wait
		}
	}
	return res
}

type PrestartStatus int 
const (
	Wait 	= iota
	Ready
	Cancel
)

type FightGame struct {
	Teams []*FightGroup
	Room_names []string
	User_names []string
	Game_id uint32
	Mode string
	User_count uint16
	Winteam int16
	Game_status uint16
	Game_port uint16
}

func (fg *FightGame) Update_names() {
	fg.Room_names = []string{}
	fg.User_names = []string{}
	for _, t := range fg.Teams {
		for _, r := range t.Rooms {
			fg.Room_names = append(fg.Room_names, r.Master)
			for _, u := range r.Users {
				fg.User_names = append(fg.User_names, u.Id)
			}
		}
	}
}

func (fg *FightGame) Check_prestart_get() bool {
	res := false
	for _, t := range fg.Teams {
		for _, r := range t.Rooms {
			res = r.Check_prestart_get()
			if (!res) {
				break
			}
		}
		if (!res) {
			break
		}
	}
	return res
}

func (fg *FightGame) Check_prestart() int {
	res := Ready
	for _, t := range fg.Teams {
		v := t.Check_prestart()
		if (v == Cancel) {
			res = Cancel
		} else if (v == Wait) {
			res = Wait
		}
	}
	return res
}

func (fg *FightGame) Set_game_id(gid uint32)  {
	for _, t := range fg.Teams {
		for _, r := range t.Rooms {
			r.Ready = 1
			for _, u := range r.Users {
				u.Game_id = gid;
			}
		}
	}
	fg.Game_id = gid
}

func (fg *FightGame) Leave_room() bool {
	for _, t := range fg.Teams {
		t.Leave_room()
	}
	return false
}

func (fg *FightGame) Clear_queue() bool {
	for _, t := range fg.Teams {
		t.Clear_queue()
	}
	return false
}

func (fg *FightGame) Ready() {
	for _, t := range fg.Teams {
		t.Ready()
	}
}