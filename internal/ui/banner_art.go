package ui

// Shared ASCII art — trailing spaces are INTENTIONAL (alignment).
// Do NOT TrimRight these lines.
var BannerArt = []string{
	` _____ _    ______          _   _____                                 `,
	`/  __ \ |   | ___ \        | | /  ___|                                `,
	`| /  \/ |__ | |_/ /__  _ __| |_\ ` + "`" + `--.  ___ __ _ _ __  _ __   ___ _ __ `,
	`| |   | '_ \|  __/ _ \| '__| __|` + "`" + `--. \/ __/ _` + "`" + ` | '_ \| '_ \ / _ \ '__|`,
	`| \__/\ | | | | | (_) | |  | |_/\__/ / (_| (_| | | | | | | |  __/ |   `,
	` \____/_| |_\_|  \___/|_|   \__\____/ \___\__,_|_| |_|_| |_|\___|_|   `,
}

var (
	dashAppName    = "ChPortScanner"
	dashAppVersion = "1.0.0"
	dashAppOwner   = "Ch4120N"
)

func SetDashboardIdentity(name, version, owner string) {
	if name != "" {
		dashAppName = name
	}
	if version != "" {
		dashAppVersion = version
	}
	if owner != "" {
		dashAppOwner = owner
	}
}