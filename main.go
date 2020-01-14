package main

import (
	"fmt"	
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"database/sql"
    _ "github.com/go-sql-driver/mysql"
	"time"
	"regexp"
	"strings"
	m "erps-go/msg"
	e "erps-go/event_room"
)

var MqttChan = make(chan m.MqttMsg)
var db *sql.DB

func generate_client_id() string {
	unix32bits := uint32(time.Now().UTC().Unix())
	uuid := fmt.Sprintf("%x",unix32bits)
	return uuid
}

func init_db() {
	db, _ = sql.Open("mysql", "erps:erpsgogo@tcp(127.0.0.1:3306)/erps")
}

func main() {
	init_db()
	//create a ClientOptions struct setting the broker address, clientid, turn
	//off trace output and set the default message handler
	e.Init(MqttChan, db)
	//go func() {
	opts := MQTT.NewClientOptions().AddBroker("114.32.129.195:1883")
	opts.SetClientID(generate_client_id())
	opts.SetDefaultPublishHandler(f)

	//create and start a client using the above ClientOptions
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {

		panic(token.Error())
	}

	c.Subscribe("member/+/send/login", 0, nil)
    c.Subscribe("member/+/send/logout", 0, nil)
    c.Subscribe("member/+/send/choose_hero", 0, nil)
    c.Subscribe("member/+/send/status", 0, nil)
    c.Subscribe("member/+/send/reconnect", 0, nil)

    c.Subscribe("room/+/send/create", 0, nil)
    c.Subscribe("room/+/send/close", 0, nil)
    c.Subscribe("room/+/send/start_queue", 0, nil)
    c.Subscribe("room/+/send/cancel_queue", 0, nil)
    c.Subscribe("room/+/send/invite", 0, nil)
    c.Subscribe("room/+/send/join", 0, nil)
    c.Subscribe("room/+/send/accept_join", 0, nil)
    c.Subscribe("room/+/send/kick", 0, nil)
    c.Subscribe("room/+/send/leave", 0, nil)
    c.Subscribe("room/+/send/prestart", 0, nil)
    c.Subscribe("room/+/send/prestart_get", 0, nil)
    c.Subscribe("room/+/send/start", 0, nil)

    c.Subscribe("game/+/send/game_close", 0, nil)
    c.Subscribe("game/+/send/game_over", 0, nil)
    c.Subscribe("game/+/send/game_info", 0, nil)
    c.Subscribe("game/+/send/start_game", 0, nil)
    c.Subscribe("game/+/send/choose", 0, nil)
    c.Subscribe("game/+/send/leave", 0, nil)
    c.Subscribe("game/+/send/exit", 0, nil)


	// Mqtt Client send message
	
	for {
		select {
		case val := <- MqttChan:
			//fmt.Printf("Channel Publish!\n")
			token := c.Publish(val.Topic, 0, false, val.Msg)
			token.Wait()
		}
	}
	//}()
	
}

 

var relogin, _ = regexp.Compile("(\\w+)/(\\w+)/send/login")
var relogout, _ = regexp.Compile("(\\w+)/(\\w+)/send/logout")
var recreate, _ = regexp.Compile("(\\w+)/(\\w+)/send/create")
var reclose, _ = regexp.Compile("(\\w+)/(\\w+)/send/close")
var restart_queue, _ = regexp.Compile("(\\w+)/(\\w+)/send/start_queue")
var recancel_queue, _ = regexp.Compile("(\\w+)/(\\w+)/send/cancel_queue")
var represtart, _ = regexp.Compile(`(\w+)/(\w+)/send/prestart`)
var represtart_get, _ = regexp.Compile(`(\w+)/(\w+)/send/prestart_get`)
var reinvite, _ = regexp.Compile("(\\w+)/(\\w+)/send/invite")
var rejoin, _ = regexp.Compile("(\\w+)/(\\w+)/send/join")
var reset, _ = regexp.Compile("reset")
var rechoose_hero, _ = regexp.Compile("(\\w+)/(\\w+)/send/choose_hero")
var releave, _ = regexp.Compile("(\\w+)/(\\w+)/send/leave")
var restart_game, _ = regexp.Compile("(\\w+)/(\\w+)/send/start_game")
var regame_over, _ = regexp.Compile("(\\w+)/(\\w+)/send/game_over")
var regame_info, _ = regexp.Compile("(\\w+)/(\\w+)/send/game_info")
var regame_close, _ = regexp.Compile("(\\w+)/(\\w+)/send/game_close")
var restatus, _ = regexp.Compile("(\\w+)/(\\w+)/send/status")
var rereconnect, _ = regexp.Compile("(\\w+)/(\\w+)/send/reconnect")


var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {	

	fmt.Println(msg.Topic())
	if (relogin.MatchString(msg.Topic())) {		
		fmt.Println("Login")
		substr := relogin.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.Login, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg

	} else if (relogout.MatchString(msg.Topic())) {
		
		substr := relogout.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.Logout, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	} else if (rereconnect.MatchString(msg.Topic())) {
		
		substr := rereconnect.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.Reconnect, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	} else if (restatus.MatchString(msg.Topic())) {
		
		substr := restatus.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.Status, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	} else if (recreate.MatchString(msg.Topic())) {
		//fmt.Printf("Create\n");
		substr := recreate.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.Create, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	} else if (reclose.MatchString(msg.Topic())) {
		//fmt.Printf("Close\n");
		substr := reclose.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.Close, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	} else if (restart_queue.MatchString(msg.Topic())) {
		//fmt.Printf("Start Queue\n");
		substr := restart_queue.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.StartQueue, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	} else if (recancel_queue.MatchString(msg.Topic())) {
		//fmt.Printf("Start Queue\n");
		substr := recancel_queue.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.CancelQueue, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	} else if (represtart.MatchString(msg.Topic())) {
		fmt.Printf("Prestart1111111\n");
		substr := represtart.FindStringSubmatch(msg.Topic())[2]
		fmt.Println(substr);
		substr1 := strings.Split(msg.Topic(), "/")[3]
		fmt.Println(substr1);
		if (substr1 == "prestart") {
			Msg:= m.ServerMsg{Event: m.PreStart, Id: substr, Msg: string(msg.Payload())}
			e.ServerChan <- Msg
		} else if (substr1 == "prestart_get") {
			Msg:= m.ServerMsg{Event: m.PreStartGet, Id: substr, Msg: string(msg.Payload())}
			e.ServerChan <- Msg
		}
	} else if (represtart_get.MatchString(msg.Topic())) {
		fmt.Printf("Prestart Get\n");
		substr := represtart_get.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.PreStartGet, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	} else if (reinvite.MatchString(msg.Topic())) {
		//fmt.Printf("Invite\n");
		substr := reinvite.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.Invite, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	} else if (rejoin.MatchString(msg.Topic())) {
		//fmt.Printf("Join\n");
		substr := rejoin.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.Join, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	} else if (releave.MatchString(msg.Topic())) {
		//fmt.Printf("leave\n");
		substr := releave.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.Leave, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	} else if (rechoose_hero.MatchString(msg.Topic())) {
		substr := rechoose_hero.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.ChooseNGHero, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	} else if (restart_game.MatchString(msg.Topic())) {
		//fmt.Printf("Start Game\n");
		substr := restart_game.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.StartGame, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	} else if (reset.MatchString(msg.Topic())) {
		//fmt.Printf("Start Game\n");
		substr := reset.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.Reset, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	}  else if (regame_over.MatchString(msg.Topic())) {
		//fmt.Printf("Game Signal\n");
		substr := regame_over.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.GameOver, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	} else if (regame_info.MatchString(msg.Topic())) {
		//fmt.Printf("Game Signal\n");
		substr := regame_info.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.GameInfo, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	} else if (regame_close.MatchString(msg.Topic())) {
		//fmt.Printf("Game Signal\n");
		substr := regame_close.FindStringSubmatch(msg.Topic())[2]
		Msg:= m.ServerMsg{Event: m.GameClose, Id: substr, Msg: string(msg.Payload())}
		e.ServerChan <- Msg
	}
}




