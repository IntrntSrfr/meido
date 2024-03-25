package customrole

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/g4s8/hexcolor"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/utils"
	"github.com/intrntsrfr/meido/pkg/utils/builders"

	"go.uber.org/zap"

	"github.com/bwmarrin/discordgo"
)

type module struct {
	*bot.ModuleBase
	db ICustomRoleDB
}

func New(b *bot.Bot, db database.DB, logger mio.Logger) bot.Module {
	logger = logger.Named("CustomRole")
	return &module{
		ModuleBase: bot.NewModule(b, "CustomRole", logger),
		db:         &CustomRoleDB{db},
	}
}

func (m *module) Hook() error {
	m.Bot.Discord.Sess.AddHandlerOnce(func(s *discordgo.Session, r *discordgo.Ready) {
		go clearDeletedRoles(m)
	})

	if err := m.RegisterApplicationCommands(
		newCustomRolesSlash(m),
	); err != nil {
		return err
	}

	if err := m.RegisterCommands(
		newSetCustomRoleCommand(m),
		newRemoveCustomRoleCommand(m),
		newMyRoleCommand(m),
		newListCustomRolesCommand(m),
	); err != nil {
		return err
	}

	return nil
}

func clearDeletedRoles(m *module) {
	exec := func(m *module) {
		m.Logger.Info("Checking for deleted roles")
		for _, g := range m.Bot.Discord.Guilds() {
			if g.Unavailable {
				continue
			}
			roles, err := m.db.GetCustomRolesByGuild(g.ID)
			if err != nil {
				continue
			}
			for _, ur := range roles {
				hasRole := false
				for _, gr := range g.Roles {
					if gr.ID == ur.RoleID {
						hasRole = true
						break
					}
				}
				if hasRole {
					continue
				}
				// delete role from guild if it no longer exists
				if err := m.db.DeleteCustomRole(ur.UID); err != nil {
					m.Logger.Error("Delete custom role failed",
						zap.Int("member roleID", ur.UID),
						zap.String("guildID", ur.GuildID),
						zap.String("roleID", ur.RoleID),
						zap.String("userID", ur.UserID))
				}
			}
		}
	}

	refreshTicker := time.NewTicker(time.Hour)
	for range refreshTicker.C {
		exec(m)
	}
}

func newSetCustomRoleCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "setcustomrole",
		Description:      "Sets or changes a custom role for a user",
		Triggers:         []string{"m?setuserrole", "m?setcustomrole"},
		Usage:            "m?setcustomrole [userID] [role]",
		Cooldown:         time.Second * 3,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionManageRoles,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 3 {
				return
			}
			targetMember, err := msg.GetMemberAtArg(1)
			if err != nil {
				_, _ = msg.Reply("Could not find that user")
				return
			}

			if targetMember.User.Bot {
				_, _ = msg.Reply("Bots dont get to join the fun")
				return
			}

			g, err := msg.Discord.Guild(msg.Message.GuildID)
			if err != nil {
				_, _ = msg.Reply(err.Error())
				return
			}

			var selectedRole *discordgo.Role
			for _, role := range g.Roles {
				if role.ID == msg.Args()[2] {
					selectedRole = role
				} else if strings.EqualFold(role.Name, strings.Join(msg.Args()[2:], " ")) {
					selectedRole = role
				}
			}

			if selectedRole == nil {
				_, _ = msg.Reply("Could not find that role")
				return
			}

			// if a custom role already exists for the user, update it to the new role
			if ur, err := m.db.GetCustomRole(msg.GuildID(), targetMember.User.ID); err == nil {
				ur.RoleID = selectedRole.ID
				if err := m.db.UpdateCustomRole(ur); err != nil {
					m.Logger.Error("Update custom role failed", zap.Error(err), zap.Any("role", ur))
					_, _ = msg.Reply("Could not set role, please try again")
					return
				}
			} else if err == sql.ErrNoRows {
				if err := m.db.CreateCustomRole(msg.GuildID(), targetMember.User.ID, selectedRole.ID); err != nil {
					m.Logger.Error("Create custom role failed", zap.Error(err), zap.String("guildID", msg.GuildID()), zap.String("roleID", selectedRole.ID), zap.String("userID", targetMember.User.ID))
					_, _ = msg.Reply("Could not set role, please try again")
					return
				}
			} else {
				m.Logger.Error("Get custom role failed", zap.Error(err))
				_, _ = msg.Reply("Could not set role, please try again")
				return
			}
			_, _ = msg.Reply(fmt.Sprintf("Set custom role for **%v** to **%v**", targetMember.User, selectedRole.Name))
		},
	}
}

func newRemoveCustomRoleCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "removecustomrole",
		Description:      "Removes a custom role that is bound to a user",
		Triggers:         []string{"m?removeuserrole", "m?removecustomrole"},
		Usage:            "m?removecustomrole [userID]",
		Cooldown:         time.Second * 3,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionManageRoles,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 2 {
				return
			}
			targetUser, err := msg.GetMemberOrUserAtArg(1)
			if err != nil {
				_, _ = msg.Reply("Could not find user")
				return
			}

			if targetUser.Bot {
				_, _ = msg.Reply("Bots dont get to join the fun")
				return
			}

			if ur, err := m.db.GetCustomRole(msg.GuildID(), targetUser.ID); err == nil {
				if err := m.db.DeleteCustomRole(ur.UID); err != nil {
					m.Logger.Error("Delete custom role failed", zap.Error(err), zap.Any("role", ur))
					_, _ = msg.Reply("Could not remove custom role, please try again")
					return
				}
				_, _ = msg.Reply(fmt.Sprintf("Removed custom role from %v", targetUser.Mention()))
			} else if err != sql.ErrNoRows {
				m.Logger.Error("Delete custom role failed", zap.Error(err), zap.Any("role", ur))
				_, _ = msg.Reply("Could not remove custom role, please try again")
			}
		},
	}
}

func newMyRoleCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "myrole",
		Description:      "Displays a users bound role, or lets the user change the name or color of their bound role",
		Triggers:         []string{"m?myrole"},
		Usage:            "m?myrole | m?myrole 123123123123 | m?myrole color c0ffee | m?myrole name jeff",
		Cooldown:         time.Second * 3,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute:          m.myroleCommand,
	}
}

func (m *module) myroleCommand(msg *discord.DiscordMessage) {
	if len(msg.Args()) < 1 {
		return
	}
	var (
		err     error
		oldRole *discordgo.Role
		target  *discordgo.Member
	)
	g, err := msg.Discord.Guild(msg.Message.GuildID)
	if err != nil {
		_, _ = msg.Reply("some error occurred")
		return
	}

	switch la := len(msg.Args()); {
	case la > 2:
		if msg.Args()[1] != "name" && msg.Args()[1] != "color" {
			return
		}
		if allow, err := msg.BotHasPermissions(discordgo.PermissionManageRoles); err != nil || !allow {
			_, _ = msg.Reply("I am missing `manage roles` permissions!")
			return
		}

		ur, err := m.db.GetCustomRole(msg.GuildID(), msg.AuthorID())
		if err != nil && err != sql.ErrNoRows {
			m.Logger.Error("error fetching user role", zap.Error(err))
			_, _ = msg.Reply("There was an issue, please try again!")
			return
		} else if err == sql.ErrNoRows {
			_, _ = msg.Reply("No custom role set")
			return
		}

		for _, role := range g.Roles {
			if role.ID == ur.RoleID {
				oldRole = role
			}
		}
		if oldRole == nil {
			_, _ = msg.Reply("Could not find custom role")
			return
		}
		topBotRole := msg.Discord.HighestRolePosition(msg.GuildID(), msg.Sess.State().User.ID)
		if oldRole.Position >= topBotRole {
			_, _ = msg.Reply("I cannot edit this role, it is above me in the role hierarchy!")
			return
		}

		if msg.Args()[1] == "name" {
			newName := strings.Join(msg.RawArgs()[2:], " ")
			if _, err = msg.Discord.Sess.GuildRoleEdit(g.ID, oldRole.ID, &discordgo.RoleParams{Name: newName}); err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				m.Logger.Error("Editing custom role name failed", zap.Error(err))
				return
			}
			embed := builders.NewEmbedBuilder().
				WithColor(oldRole.Color).
				WithDescription(fmt.Sprintf("Role name changed from %v to %v", oldRole.Name, newName))
			_, _ = msg.ReplyEmbed(embed.Build())

		} else if msg.Args()[1] == "color" {
			clrStr := msg.Args()[2]
			if !strings.HasPrefix(clrStr, "#") {
				clrStr = "#" + strings.TrimSpace(clrStr)
			}
			clr, err := hexcolor.Parse(clrStr)
			if err != nil {
				_, _ = msg.ReplyEmbed(&discordgo.MessageEmbed{Description: "Invalid color code", Color: utils.ColorCritical})
				return
			}

			clrInt := int(clr.R)<<16 + int(clr.G)<<8 + int(clr.B)
			if _, err = msg.Discord.Sess.GuildRoleEdit(g.ID, oldRole.ID, &discordgo.RoleParams{Color: &clrInt}); err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				m.Logger.Error("Editing custom role color failed", zap.Error(err))
				return
			}
			embed := builders.NewEmbedBuilder().
				WithColor(clrInt).
				WithDescription(fmt.Sprintf("Color changed from #%06X to #%06X", oldRole.Color, clrInt))
			_, _ = msg.ReplyEmbed(embed.Build())
		}
		return
	case la == 1:
		target = msg.Member()
	case la == 2:
		if target, err = msg.GetMemberAtArg(1); err != nil {
			_, _ = msg.Reply("Could not find that user")
			return
		}
	default:
		return
	}
	if target == nil {
		return
	}

	ur, err := m.db.GetCustomRole(msg.GuildID(), target.User.ID)
	if err != nil && err != sql.ErrNoRows {
		_, _ = msg.Reply("there was an error, please try again")
		m.Logger.Error("Fetching custom role failed", zap.Error(err))
		return
	} else if err == sql.ErrNoRows {
		_, _ = msg.Reply("No custom role set.")
		return
	}

	var customRole *discordgo.Role
	for i := range g.Roles {
		role := g.Roles[i]
		if role.ID == ur.RoleID {
			customRole = role
		}
	}
	if customRole == nil {
		_, _ = msg.Reply("The custom role is broken, wait for someone to fix it or try setting a new custom role")
		return
	}

	embed := builders.NewEmbedBuilder().
		WithTitle(fmt.Sprintf("Custom role for %v", target.User.String())).
		WithColor(customRole.Color).
		AddField("Name", customRole.Name, true).
		AddField("Color", fmt.Sprintf("#%06X", customRole.Color), true)
	_, _ = msg.ReplyEmbed(embed.Build())
}

func newCustomRolesSlash(m *module) *bot.ModuleApplicationCommand {
	bld := bot.NewModuleApplicationCommandBuilder(m, "customroles").
		Type(discordgo.ChatApplicationCommand).
		Description("Add, remove, or list custom roles").
		Permissions(discordgo.PermissionManageRoles).
		NoDM().
		Cooldown(time.Second*3, bot.CooldownScopeChannel).
		AddOption(builders.NewSubCommandBuilder("add", "Add a custom role").
			AddOption(&discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The user to add the custom role to",
				Required:    true,
			}).
			AddOption(&discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "role",
				Description: "The role to add",
				Required:    true,
			}).Build()).
		AddOption(builders.NewSubCommandBuilder("remove", "Remove a custom role").
			AddOption(&discordgo.ApplicationCommandOption{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The user to remove the custom role from",
				Required:    true,
			}).Build()).
		AddOption(builders.NewSubCommandBuilder("list", "List all custom roles").Build())

	exec := func(dac *discord.DiscordApplicationCommand) {
		switch dac.SubCommand() {
		case "add":
			addCustomRole(m, dac)
		case "remove":
			removeCustomRole(m, dac)
		case "list":
			listCustomRoles(m, dac)
		}
	}

	return bld.Execute(exec).Build()
}

func addCustomRole(m *module, dac *discord.DiscordApplicationCommand) {
	userOpt, userOk := dac.Options("add:user")
	roleOpt, roleOk := dac.Options("add:role")
	if !userOk || !roleOk {
		_ = dac.RespondEphemeral("Missing user or role")
		return
	}

	user := userOpt.UserValue(dac.Sess.Real())
	role := roleOpt.RoleValue(dac.Sess.Real(), dac.GuildID())
	if user == nil || role == nil {
		_ = dac.RespondEphemeral("Could not find user or role")
		return
	}

	if user.Bot {
		_ = dac.RespondEphemeral("Bots dont get to join the fun")
		return
	}

	if _, err := dac.Discord.Member(dac.GuildID(), user.ID); err != nil {
		_ = dac.RespondEphemeral("That user is not in the server")
		return
	}

	// if a custom role already exists for the user, update it to the new role
	if ur, err := m.db.GetCustomRole(dac.GuildID(), user.ID); err == nil {
		ur.RoleID = role.ID
		if err := m.db.UpdateCustomRole(ur); err != nil {
			m.Logger.Error("Update custom role failed", zap.Error(err), zap.Any("role", ur))
			_ = dac.RespondEphemeral("Could not set role, please try again")
			return
		}
	} else if err == sql.ErrNoRows {
		if err := m.db.CreateCustomRole(dac.GuildID(), user.ID, role.ID); err != nil {
			m.Logger.Error("Create custom role failed", zap.Error(err), zap.String("guildID", dac.GuildID()), zap.String("roleID", role.ID), zap.String("userID", user.ID))
			_ = dac.RespondEphemeral("Could not set role, please try again")
			return
		}
	} else {
		m.Logger.Error("Get custom role failed", zap.Error(err))
		_ = dac.RespondEphemeral("Could not set role, please try again")
		return
	}
	_ = dac.Respond(fmt.Sprintf("Set custom role for **%v** to **%v**", user, role.Name))
}

func removeCustomRole(m *module, dac *discord.DiscordApplicationCommand) {
	userOpt, userOk := dac.Options("remove:user")
	if !userOk {
		_ = dac.RespondEphemeral("Missing user")
		return
	}

	user := userOpt.UserValue(dac.Sess.Real())
	if user == nil {
		_ = dac.RespondEphemeral("Could not find user")
		return
	}

	if user.Bot {
		_ = dac.RespondEphemeral("Bots dont get to join the fun")
		return
	}

	if ur, err := m.db.GetCustomRole(dac.Interaction.GuildID, user.ID); err == nil {
		if err := m.db.DeleteCustomRole(ur.UID); err != nil {
			m.Logger.Error("Delete custom role failed", zap.Error(err), zap.Any("role", ur))
			_ = dac.RespondEphemeral("Could not remove custom role, please try again")
			return
		}
		_ = dac.Respond(fmt.Sprintf("Removed custom role from **%v**", user))
		return
	}
}

func listCustomRoles(m *module, dac *discord.DiscordApplicationCommand) {
	roles, err := m.db.GetCustomRolesByGuild(dac.Interaction.GuildID)
	if err != nil {
		m.Logger.Error("Error fetching custom roles", zap.Error(err))
		_ = dac.RespondEphemeral("There was an issue, please try again!")
		return
	}

	g, err := dac.Sess.Guild(dac.Interaction.GuildID)
	if err != nil {
		_ = dac.RespondEphemeral("There was an issue, please try again!")
		return
	}

	roleList := generateRoleList(dac.Discord, g, roles)
	data := &discordgo.InteractionResponseData{Content: roleList}
	if len(roleList) > 10 {
		data.Files = []*discordgo.File{{
			Name:   "roles.txt",
			Reader: bytes.NewBufferString(roleList),
		}}
		data.Content = ""
	}
	_ = dac.RespondComplex(data, discordgo.InteractionResponseChannelMessageWithSource)
}

func newListCustomRolesCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "listcustomroles",
		Description:      "Lists all custom roles in the guild and who they belong to",
		Triggers:         []string{"m?listuserroles", "m?listcustomroles"},
		Usage:            "m?listcustomroles",
		Cooldown:         time.Second * 30,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionManageRoles,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			roles, err := m.db.GetCustomRolesByGuild(msg.GuildID())
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				return
			}

			g, err := msg.Discord.Guild(msg.Message.GuildID)
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				return
			}

			roleList := generateRoleList(msg.Discord, g, roles)
			data := &discordgo.MessageSend{Content: roleList}
			if len(roleList) > 1024 {
				data.File = &discordgo.File{
					Name:   "roles.txt",
					Reader: bytes.NewBufferString(roleList),
				}
				data.Content = ""
			}
			_, _ = msg.ReplyComplex(data)
		},
	}
}

func generateRoleList(d *discord.Discord, g *discordgo.Guild, roles []*CustomRole) string {
	builder := strings.Builder{}
	w := tabwriter.NewWriter(&builder, 0, 0, 4, ' ', 0)
	builder.WriteString(fmt.Sprintf("Custom roles in %v | Amount: %v\n\n", g.Name, len(roles)))
	fmt.Fprintln(w, "Role Name\tRole ID\tBelongs to\tUser ID")
	for _, ur := range roles {
		role, err := d.Role(g.ID, ur.RoleID)
		if err != nil {
			continue
		}
		mem, err := d.Member(g.ID, ur.UserID)
		if err != nil {
			fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", role.Name, role.ID, "[Not in server]", ur.UserID)
		} else {
			fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", role.Name, role.ID, mem.User, mem.User.ID)
		}
	}
	w.Flush()
	return builder.String()
}
