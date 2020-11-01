package meidov2

import "fmt"

type Config struct {
	Token string `json:"token"`
}

type Bot struct {
	Discord    *Discord
	Config     *Config
	Mods       map[string]Mod
	commandLog chan *DiscordMessage
}

func NewBot(config *Config) *Bot {
	d := NewDiscord(config.Token)

	fmt.Println("new bot")

	return &Bot{
		Discord:    d,
		Config:     config,
		Mods:       make(map[string]Mod),
		commandLog: make(chan *DiscordMessage, 256),
	}
}

func (b *Bot) Open() error {
	msgChan, err := b.Discord.Open()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("open and run")

	go b.listen(msgChan)
	go b.logCommands()
	return nil
}

func (b *Bot) Run() error {
	return b.Discord.Run()
}

func (b *Bot) Close() {
	b.Discord.Client.Disconnect()
}

func (b *Bot) RegisterMod(mod Mod, name string) {
	fmt.Println(fmt.Sprintf("registering mod '%s'", name))
	err := mod.Hook(b, b.commandLog)
	if err != nil {
		fmt.Println("could not attach mod:", name)
		return
	}
	b.Mods[name] = mod
}

func (b *Bot) listen(msg <-chan *DiscordMessage) {
	for {
		select {
		case m := <-msg:
			if m.Type != MessageTypeCreate || m.Message.Author.Bot {
				continue
			}
			for _, mod := range b.Mods {
				mod.Message(m)
			}
		default:
			continue
		}
	}
}

func (b *Bot) logCommands() {
	for {
		select {
		case e := <-b.commandLog:
			fmt.Println(e.Message.Author.String(), e.Message.Content, e.TimeReceived.String())
		}
	}
}
