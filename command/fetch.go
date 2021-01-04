package command

import (
	"encoding/json"
	"log"
	"strings"
	"trup/db"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/jackc/pgx"
)

const setFetchHelp = "Run without arguments to see instructions."

func setFetch(ctx *Context, args []string) {
	lines := strings.Split(ctx.Message.Content, "\n")
	if len(lines) < 2 && len(ctx.Message.Attachments) == 0 {
		ctx.Reply("run this: `curl -s https://raw.githubusercontent.com/unixporn/trup/prod/fetcher.sh | sh` and follow the instructions. It's recommended that you download and read the script before running it, as piping to curl isn't always the safest practice. (<https://blog.dijit.sh/don-t-pipe-curl-to-bash>)\n > NOTE: use `!setfetch update` to update individual values\n > NOTE: If you're trying to manually change a value, it needs a newline. Also note that !git and !dotfiles are different commands.")

		return
	}

	var data db.SysinfoData

	if len(args) >= 2 && args[1] == "update" {
		sysinfo, err := db.GetSysinfo(ctx.Message.Author.ID)
		if err != nil {
			if err.Error() == pgx.ErrNoRows.Error() {
				ctx.Reply("Cannot update fetch; No existing fetch data found")
			} else {
				ctx.ReportError("Failed to get existing fetch data", err)
			}

			return
		}
		data = sysinfo.Info
	}

	m := map[string]*string{
		"Distro":           &data.Distro,
		"Kernel":           &data.Kernel,
		"Terminal":         &data.Terminal,
		"Editor":           &data.Editor,
		"DE/WM":            &data.DeWm,
		"Bar":              &data.Bar,
		"Resolution":       &data.Resolution,
		"Display Protocol": &data.DisplayProtocol,
		"GTK3 Theme":       &data.Gtk3Theme,
		"GTK Icon Theme":   &data.GtkIconTheme,
		"CPU":              &data.Cpu,
		"GPU":              &data.Gpu,
	}

	for i := 1; i < len(lines); i++ {
		kI := strings.Index(lines[i], ":")
		if kI == -1 {
			continue
		}

		key := lines[i][:kI]
		value := strings.TrimSpace(lines[i][kI+1:])

		if isValidURL(lines[i]) {
			data.Image = lines[i]

			continue
		}

		if addr, found := m[key]; found {
			*addr = value

			continue
		}

		switch key {
		case "Memory":
			b, err := humanize.ParseBytes(value)
			if err != nil {
				ctx.Reply("Failed to parse Max RAM")
				return
			}

			data.Memory = b
		default:
			ctx.Reply("key '" + key + "' is not valid")

			return
		}
	}

	for _, a := range ctx.Message.Embeds {
		if a.Type == "image" {
			data.Image = a.URL
			break
		}
	}

	for _, a := range ctx.Message.Attachments {
		if a.Width > 0 {
			data.Image = a.URL
			break
		}
	}

	info := db.NewSysinfo(ctx.Message.Author.ID, data)
	err := info.Save()
	if err != nil {
		ctx.Reply("Failed to save. Error: " + err.Error())
		return
	}
	ctx.Reply("success. You can now run !fetch")
}

const fetchUsage = "fetch [user]"

func doFetch(ctx *Context, user *discordgo.User) {
	const inline = true
	embed := discordgo.MessageEmbed{
		Title:  "Fetch " + user.Username + "#" + user.Discriminator,
		Fields: []*discordgo.MessageEmbedField{},
	}

	profile, err := db.GetProfile(user.ID)
	if err != nil && err.Error() != pgx.ErrNoRows.Error() {
		whose := "your"
		if user.ID != ctx.Message.Author.ID {
			whose = "the user's"
		}
		ctx.ReportError("Failed to fetch "+whose+" profile.", err)
	}
	profileFields := []*discordgo.MessageEmbedField{}

	if err == nil {
		if profile.Description != "" {
			embed.Description = profile.Description
		}

		if profile.Git != "" {
			profileFields = append(profileFields, &discordgo.MessageEmbedField{
				Name:   "Git",
				Value:  profile.Git,
				Inline: inline,
			})
		}

		if profile.Dotfiles != "" {
			profileFields = append(profileFields, &discordgo.MessageEmbedField{
				Name:   "Dotfiles",
				Value:  profile.Dotfiles,
				Inline: inline,
			})
		}
	}

	somethingToPost := len(embed.Fields) > 0 || len(profileFields) > 0 || embed.Description != ""

	info, err := db.GetSysinfo(user.ID)
	if err != nil {
		// err == pgx.ErrNoRows doesn't work, not sure why
		if err.Error() == pgx.ErrNoRows.Error() {
			if somethingToPost {
				goto sysinfoEnd
			}

			message := "that user hasn't set their fetch information. You can ask them to run !setfetch"
			if user.ID == ctx.Message.Author.ID {
				message = "you first need to set your information with !setfetch"
			}

			ctx.Reply(message)
			return
		}

		ctx.Reply("failed to find the user's info. Error: " + err.Error())
		return
	}

	embed.Color = ctx.Session.State.UserColor(user.ID, ctx.Message.ChannelID)
	if info.Info.Distro != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: getDistroImage(info.Info.Distro),
		}
	}

	if info.Info.Distro != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Distro",
			Value:  info.Info.Distro,
			Inline: inline,
		})
	}

	if info.Info.Kernel != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Kernel",
			Value:  info.Info.Kernel,
			Inline: inline,
		})
	}

	if info.Info.Terminal != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Terminal",
			Value:  info.Info.Terminal,
			Inline: inline,
		})
	}

	if info.Info.Editor != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Editor",
			Value:  info.Info.Editor,
			Inline: inline,
		})
	}

	if info.Info.DeWm != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "DE/WM",
			Value:  info.Info.DeWm,
			Inline: inline,
		})
	}

	if info.Info.Bar != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Bar",
			Value:  info.Info.Bar,
			Inline: inline,
		})
	}

	if info.Info.Resolution != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Resolution",
			Value:  info.Info.Resolution,
			Inline: inline,
		})
	}

	if info.Info.DisplayProtocol != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Display Protocol",
			Value:  info.Info.DisplayProtocol,
			Inline: inline,
		})
	}

	if info.Info.Gtk3Theme != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "GTK3 Theme",
			Value:  info.Info.Gtk3Theme,
			Inline: inline,
		})
	}

	if info.Info.GtkIconTheme != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "GTK Icon Theme",
			Value:  info.Info.GtkIconTheme,
			Inline: inline,
		})
	}

	if info.Info.Cpu != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "CPU",
			Value:  info.Info.Cpu,
			Inline: inline,
		})
	}

	if info.Info.Gpu != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "GPU",
			Value:  info.Info.Gpu,
			Inline: inline,
		})
	}

	if info.Info.Memory != 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Memory",
			Value:  humanize.Bytes(info.Info.Memory),
			Inline: inline,
		})
	}

	if info.Info.Image != "" {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: info.Info.Image,
		}
	}

	if !info.ModifyDate.IsZero() {
		const dateFormat = "2006-01-02T15:04:05.0000Z"
		embed.Timestamp = info.ModifyDate.UTC().Format(dateFormat)
	}

sysinfoEnd:
	embed.Fields = append(embed.Fields, profileFields...)

	_, err = ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &embed)
	if err != nil {
		var retry bool
		if restErr, ok := err.(*discordgo.RESTError); ok {
			var embedErr embedError
			if err := json.Unmarshal(restErr.ResponseBody, &embedErr); err != nil {
				log.Printf("Failed to unmarshal RESTError to embedError, err: %s\n", err)
				return
			}
			for _, field := range embedErr.Embed {
				if field == "image" {
					embed.Image = nil
					retry = true
				}
			}
		}

		if retry {
			if _, err = ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &embed); err != nil {
				log.Println("Failed to send channel embed: " + err.Error())
			}
		}
	}
}

func fetch(ctx *Context, args []string) {
	if len(args) < 2 {
		doFetch(ctx, ctx.Message.Author)
	} else {
		err := ctx.requestUserByName(false, strings.Join(args[1:], " "), func(member *discordgo.Member) error {
			doFetch(ctx, member.User)
			return nil
		})
		if err != nil {
			ctx.ReportError("Failed to find the user.", err)
			return
		}
	}
}

func getDistroImage(name string) string {
	name = strings.ToLower(name)

	for _, d := range distroImages {
		if strings.HasPrefix(name, d.name) {
			return d.image
		}
	}

	return ""
}

var distroImages = []struct {
	name  string
	image string
}{
	{name: "nixos", image: "https://nixos.org/logo/nixos-hires.png"},
	{name: "arch", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/7/74/Arch_Linux_logo.svg/250px-Arch_Linux_logo.svg.png"},
	{name: "archbang", image: "https://upload.wikimedia.org/wikipedia/commons/2/2c/ArchBangLogo.png"},
	{name: "archlabs", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/7/73/Default_desktop.png/300px-Default_desktop.png"},
	{name: "artix", image: "https://artixlinux.org/img/artix-logo.png"},
	{name: "alpine", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/e/e6/Alpine_Linux.svg/250px-Alpine_Linux.svg.png"},
	{name: "alt", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/4/44/AltLinux500Desktop.png/250px-AltLinux500Desktop.png"},
	{name: "antergos", image: "https://upload.wikimedia.org/wikipedia/en/thumb/9/93/Antergos_logo_github.png/150px-Antergos_logo_github.png"},
	{name: "backbox", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b2/BackBox_4.4_Screenshot.png/250px-BackBox_4.4_Screenshot.png"},
	{name: "boss", image: "https://upload.wikimedia.org/wikipedia/en/f/f2/Bharat_Operating_System_Solutions_logo%2C_Sept_2015.png"},
	{name: "bodhi", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/f/fc/Bodhi_Linux_Logo.png/250px-Bodhi_Linux_Logo.png"},
	{name: "calculate", image: "https://upload.wikimedia.org/wikipedia/commons/3/3a/CalculateLinux-transparent-90x52.png"},
	{name: "caos", image: "https://upload.wikimedia.org/wikipedia/en/4/4b/CAos_Linux_logo.png"},
	{name: "centos", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/b/bf/Centos-logo-light.svg/300px-Centos-logo-light.svg.png"},
	{name: "cub", image: "https://upload.wikimedia.org/wikipedia/commons/d/d8/CubLinux100.png"},
	{name: "debian", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/4/4a/Debian-OpenLogo.svg/100px-Debian-OpenLogo.svg.png"},
	{name: "deepin", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/f/f5/Deepin_logo.svg/60px-Deepin_logo.svg.png"},
	{name: "elementary", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/8/83/Elementary_OS_logo.svg/300px-Elementary_OS_logo.svg.png"},
	{name: "emmabuntüs", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/9/95/Emmabuntus_DE3_En.png/150px-Emmabuntus_DE3_En.png"},
	{name: "engarde", image: "https://upload.wikimedia.org/wikipedia/en/7/74/Engarde_Logo.png"},
	{name: "euleros", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/e/e1/Operating_system_placement.svg/24px-Operating_system_placement.svg.png"},
	{name: "fedora", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/0/09/Fedora_logo_and_wordmark.svg/250px-Fedora_logo_and_wordmark.svg.png"},
	{name: "fermi", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/a/a5/Fermi_Linux_logo.svg/80px-Fermi_Linux_logo.svg.png"},
	{name: "finnix", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/5/52/Finnix-72pt-72dpi.png/100px-Finnix-72pt-72dpi.png"},
	{name: "foresight", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/4/48/Foresight_Linux_Logo_2.png/200px-Foresight_Linux_Logo_2.png"},
	{name: "frugalware", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/3/3c/Frugalware_linux_logo.svg/250px-Frugalware_linux_logo.svg.png"},
	{name: "fuduntu", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/2/2e/Fuduntu-logo.png/100px-Fuduntu-logo.png"},
	{name: "geckolinux", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/3/35/Tux.svg/35px-Tux.svg.png"},
	{name: "gentoo", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/4/48/Gentoo_Linux_logo_matte.svg/100px-Gentoo_Linux_logo_matte.svg.png"},
	{name: "hyperbola", image: "https://www.hyperbola.info/img/devs/silhouette.png"},
	{name: "instantos", image: "https://media.githubusercontent.com/media/instantOS/instantLOGO/master/png/light.png"},
	{name: "kali", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/4/4b/Kali_Linux_2.0_wordmark.svg/131px-Kali_Linux_2.0_wordmark.svg.png"},
	{name: "kanotix", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/c/c4/Kanotix-hellfire.png/300px-Kanotix-hellfire.png"},
	{name: "kaos", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/2/2c/KaOS_201603.png/300px-KaOS_201603.png"},
	{name: "kde neon", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/f/f7/Neon-logo.svg/100px-Neon-logo.svg.png"},
	{name: "kororā", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/6/6e/Korora_logo.png/250px-Korora_logo.png"},
	{name: "kubuntu", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/7/76/Kubuntu_logo_and_wordmark.svg/250px-Kubuntu_logo_and_wordmark.svg.png"},
	{name: "kwort", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/4/49/2019-11-24-121414_1280x1024_scrot.png/300px-2019-11-24-121414_1280x1024_scrot.png"},
	{name: "linux lite", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/7/79/Linux_Lite_Simple_Fast_Free_logo.png/250px-Linux_Lite_Simple_Fast_Free_logo.png"},
	{name: "linux mint", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/5/5c/Linux_Mint_Official_Logo.svg/250px-Linux_Mint_Official_Logo.svg.png"},
	{name: "lunar linux", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/a/a1/Lunar_Linux_logo.png/200px-Lunar_Linux_logo.png"},
	{name: "macos", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/2/21/MacOS_wordmark_%282017%29.svg/200px-MacOS_wordmark_%282017%29.svg.png"},
	{name: "mageia", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/9/93/Mageia_logo.svg/250px-Mageia_logo.svg.png"},
	{name: "mandriva", image: "https://upload.wikimedia.org/wikipedia/en/thumb/3/34/Mandriva-Logo.svg/200px-Mandriva-Logo.svg.png"},
	{name: "manjaro", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/a/a5/Manjaro_logo_text.png/250px-Manjaro_logo_text.png"},
	{name: "simplymepis", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/f/fc/MEPIS_logo.svg/100px-MEPIS_logo.svg.png"},
	{name: "mx linux", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/d/d4/MX_Linux_logo.svg/100px-MX_Linux_logo.svg.png"},
	{name: "openmandriva lx", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/6/60/Oma-logo-22042013_300pp.png/70px-Oma-logo-22042013_300pp.png"},
	{name: "opensuse", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/d/d0/OpenSUSE_Logo.svg/128px-OpenSUSE_Logo.svg.png"},
	{name: "oracle", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/5/50/Oracle_logo.svg/200px-Oracle_logo.svg.png"},
	{name: "parted magic", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/1/11/Parted_Magic_2014_04_28.png/300px-Parted_Magic_2014_04_28.png"},
	{name: "pclinuxos", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/8/89/PCLinuxOS_logo.svg/80px-PCLinuxOS_logo.svg.png"},
	{name: "pinguy os", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/7/7a/Pinguy-os-desktop-12-04.png/300px-Pinguy-os-desktop-12-04.png"},
	{name: "pop!_os", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/0/02/Pop_OS-Logo-nobg.png/125px-Pop_OS-Logo-nobg.png"},
	{name: "qubes os", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/6/61/Qubes_OS_Logo.svg/120px-Qubes_OS_Logo.svg.png"},
	{name: "red flag", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/6/6c/RedFlag_Linux-Logo.jpg/180px-RedFlag_Linux-Logo.jpg"},
	{name: "red hat enterprise", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/4/46/RHEL_8_Desktop.png/300px-RHEL_8_Desktop.png"},
	{name: "rosa linux", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/2/25/Logo_ROSA.jpg/250px-Logo_ROSA.jpg"},
	{name: "russian fedora remix project", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/0/08/Rfremix_Logo9.png/300px-Rfremix_Logo9.png"},
	{name: "sabayon", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/3/3d/Sabayon_5.4_logo.svg/70px-Sabayon_5.4_logo.svg.png"},
	{name: "sailfish os", image: "https://upload.wikimedia.org/wikipedia/en/thumb/d/d3/Sailfish_logo.svg/250px-Sailfish_logo.svg.png"},
	{name: "scientific", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b1/Scientific_Linux_logo_and_wordmark.svg/80px-Scientific_Linux_logo_and_wordmark.svg.png"},
	{name: "slackware", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/2/22/Slackware_logo_from_the_official_Slackware_site.svg/250px-Slackware_logo_from_the_official_Slackware_site.svg.png"},
	{name: "solus", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/f/ff/Solus.svg/100px-Solus.svg.png"},
	{name: "solydxk", image: "https://upload.wikimedia.org/wikipedia/en/d/df/SolydXK_logo_small.png"},
	{name: "sparkylinux", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/1/16/SparkyLinux-logo-200px.png/110px-SparkyLinux-logo-200px.png"},
	{name: "suse linux enterprise desktop", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/5/59/SLED_15_Default_Desktop.png/300px-SLED_15_Default_Desktop.png"},
	{name: "suse linux enterprise server", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/f/ff/SUSE_Linux_Enterprise_Server_11_installation_DVD_20100429.jpg/300px-SUSE_Linux_Enterprise_Server_11_installation_DVD_20100429.jpg"},
	{name: "turbolinux", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/f/f1/Turbolinux.png/250px-Turbolinux.png"},
	{name: "turnkey linux virtual appliance library", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/a/a1/Image-webmin3.png/300px-Image-webmin3.png"},
	{name: "ubuntu budgie", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/2/2e/UbuntuBudgie-Wordmark.svg/250px-UbuntuBudgie-Wordmark.svg.png"},
	{name: "ubuntu gnome", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/4/41/Ubuntu_GNOME_logo.svg/250px-Ubuntu_GNOME_logo.svg.png"},
	{name: "ubuntu mate", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/5/53/Ubuntu_MATE_logo.svg/250px-Ubuntu_MATE_logo.svg.png"},
	{name: "ubuntu", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/3/3a/Logo-ubuntu_no%28r%29-black_orange-hex.svg/250px-Logo-ubuntu_no%28r%29-black_orange-hex.svg.png"},
	{name: "univention corporate server", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b6/Univention_Corporate_Server_Logo.png/250px-Univention_Corporate_Server_Logo.png"},
	{name: "uruk", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/3/39/Logo_of_Uruk_Project.svg/250px-Logo_of_Uruk_Project.svg.png"},
	{name: "vine", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/3/35/Tux.svg/35px-Tux.svg.png"},
	{name: "void", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/0/02/Void_Linux_logo.svg/200px-Void_Linux_logo.svg.png"},
	{name: "whonix", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/7/75/Whonix_Logo.png/200px-Whonix_Logo.png"},
	{name: "xubuntu", image: "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b0/Xubuntu_logo_and_wordmark.svg/200px-Xubuntu_logo_and_wordmark.svg.png"},
}
