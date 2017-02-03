package bot


import (
    "time"
    "fmt"
    
    "github.com/tucnak/telebot"
)

type TelegramBot struct{
	Bot *telebot.Bot
    Token string
    Name string // bot name, will be the tweetssaver info to know if this instance is OK
    Started bool
}

var instance *TelegramBot = nil

func CreateInstance(token, name string) *TelegramBot {
    instance = &TelegramBot{Token:token, Name:name, Started:false}
    bot, err := telebot.NewBot(token)
    if err != nil {
        return nil
    }

    instance.Started = true
    instance.Bot = bot
    go listenMessages()

    return instance
}

func GetInstance() *TelegramBot {
    return instance
}

func RefreshSession(){
    CreateInstance(instance.Token, instance.Name)
}


func listenMessages(){
    myBot := GetInstance()
    messages := make(chan telebot.Message)
    myBot.Bot.Listen(messages, 1*time.Second)

    for message := range messages {
        callToMe := fmt.Sprintf("/%s", myBot.Name)
        userID := fmt.Sprintf("%d", message.Sender.ID)
        if message.Text == callToMe {
            myBot.Bot.SendMessage(message.Chat, "Hello, here I am! All is fine! :)", nil)
        }else if message.Text == "/info" {
            myBot.Bot.SendMessage(message.Chat, "Your ID is " + userID, nil)
        }else{
            r := fmt.Sprintf("[%s] All seems fine here! :)", myBot.Name)
            myBot.Bot.SendMessage(message.Chat, r, nil)
        }
    }
}

func SendMessage(to int64, message string){
    myBot := GetInstance()

    chat := telebot.Chat{ID: to}
    myBot.Bot.SendMessage(chat, message, nil)
}

