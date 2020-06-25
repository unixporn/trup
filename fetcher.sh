#!/bin/sh

kernel="$(uname -s)"

print() {
cat << EOF
Copy and paste the command below in the server.
You can also attach an image to the message, be it your screenshot or wallpaper.

!setfetch
Distro: $NAME $ver
Kernel: $(uname -sr)
Terminal:$term
$([ "$wm" ] && echo "DE/WM: $wm" || echo "Display protocol: $displayprot")
Editor: $EDITOR
GTK3 Theme: $theme
GTK Icon Theme: $icons
CPU: $cpu
GPU: $gpu
Memory: $ram
EOF
}

if [ "$kernel" = "Linux" ]; then
	# get distro
	# name is saved in the $NAME variable
	. "/etc/os-release"

	# get display protocol
	[ "$DISPLAY" ] && displayprot="x11"
	[ "$WAYLAND_DISPLAY" ] && displayprot="wayland"
	# fallback to tty if none is detected
	[ ! "$displayprot" ] && displayprot="tty"

	# get gtk theme
	gtkrc="${XDG_CONFIG_HOME:-$HOME/.config}/gtk-3.0/settings.ini"
	theme="$(test -f "$gtkrc" && awk -F'=' '/gtk-theme-name/ {print $2} ' "$gtkrc")" &&
	icons="$(awk -F'=' '/gtk-icon-theme-name/ {print $2} ' "$gtkrc")"

	# TODO: Support for detecting Wayland Compositors
	# check for wm on X11
	[ "$DISPLAY" ] && {
		# for standard WMs
		command -v xprop >/dev/null 2>&1 && {
			id=$(xprop -root -notype _NET_SUPPORTING_WM_CHECK)
			id=${id##* }
			wm="$(xprop -id "$id" -notype -len 100 -f _NET_WM_NAME 8t | \
				grep WM_NAME | cut -d' ' -f 3 | tr -d '"')"
		}

		# Fallback for non-EWMH WMs
		[ "$wm" ] ||
			wm=$(ps -e | grep -m 1 -o \
				-e "[s]owm" \
				-e "[c]atwm" \
				-e "[f]vwm" \
				-e "[d]wm" \
				-e "[2]bwm" \
				-e "[m]onsterwm" \
				-e "[t]inywm" \
				-e "[x]monad")
	}

	# hardware
	cpu="$(awk -F': ' '/model name\t: /{print $2;exit} ' "/proc/cpuinfo")"
	ram="$(awk '/[^0-9]* / {print $2" "$3;exit} ' "/proc/meminfo")"

        # gpu
        if
          [ "$(lspci | grep -i vga | grep -i nvidia | awk '{print $5}')" ]
        then
	  gpu="$(lspci | grep -iE 'vga|display|3d' | grep -iE nvidia | awk '{print $5 " " $8 " " $9 " " $10 " " $11}'| sed -e 's/\[\|\]//g')"
        elif
          [ "$(lspci | grep -iE 'vga|display|3d' | grep -i amd)" ]
        then
	  gpu="$(lspci | grep -iE 'vga|display|3d' | grep -iE amd | awk '{print $5 " " $9 " " $11 " " $12}' | sed -e 's/\[\|\]//g')"
        elif
          [ "$(lspci | grep -iE 'vga|display|3d' | grep -i radeon)" ]
        then
	  gpu="$(lspci | grep -iE 'vga|display|3d' | grep -iE radeon | awk '{print $5 " " $9 " " $11 " " $12}' | sed -e 's/\[\|\]//g')"
        elif
          [ "$(lspci | grep -iE 'vga|display|3d' | grep -i rx)" ]
        then
          gpu="$(lspci | grep -iE 'vga|display|3d' | grep -iE rx | awk '{print $5 " " $9 " " $11 " " $12}' | sed -e 's/\[\|\]//g')"
        elif
          [ "$(lspci | grep -iE 'vga|display|3d' | grep -i xt)" ]
        then
          gpu="$(lspci | grep -iE 'vga|display|3d' | grep -iE xt | awk '{print $5 " " $9 " " $11 " " $12}' | sed -e 's/\[\|\]//g')"
        elif
          [ "$(lspci | grep -iE 'vga|display|3d' | grep -i vega)" ]
        then
          gpu="$(lspci | grep -iE 'vga|display|3d' | grep -iE vega | awk '{print $5 " " $9 " " $11 " " $12}' | sed -e 's/\[\|\]//g')"
        elif
          [ "$(lspci | grep -i vga | grep -i intel | awk '{print $5}')" ]
        then
	  gpu="$(lspci | grep -E 'VGA|Display' | grep -iE intel | awk '{print $5 " " $8 " " $9 " " $10 " " $11}' | cut -d'(' -f1)"
	else
          gpu=" "
        fi

	# editor, remove the file path
	[ "$EDITOR" ] && EDITOR="${EDITOR##*/}"

	# terminal, remove declaration of color support from the name
	term=$(ps -e | grep -m 1 -o \
		-e " alacritty$" \
		-e " kitty$" \
		-e " xterm$" \
		-e " urxvt$" \
		-e " xfce4-terminal$" \
		-e " gnome-terminal$" \
		-e " mate-terminal$" \
		-e " cool-retro-term$" \
		-e " konsole$" \
		-e " termite$" \
		-e " rxvt$" \
		-e " tilix$" \
		-e " sakura$" \
		-e " terminator$" \
		-e " qterminal$" \
		-e " termonad$" \
		-e " lxterminal$" \
		-e " st$" \
		-e " xst$" \
		-e " tilda$")

	print
elif [ "$kernel"  = "Darwin" ]; then
	NAME="macOS"

	# get MacOS version
	# example output: <string>10.15.4</string>
	ver="$(awk '/ProductVersion/{getline; print}' /System/Library/CoreServices/SystemVersion.plist)"
	# remove <string>
	ver="${ver#*>}"
	# remove </string>
	ver="${ver%<*}"

	# get WM
	wm="$(ps -e | grep -o \
		-e "[S]pectacle" \
		-e "[A]methyst" \
		-e "[k]wm" \
		-e "[c]hun[k]wm" \
		-e "[y]abai" \
		-e "[R]ectangle" | head -n1)"

	# if the current WM isn't on this list, assume default DE
	wm="${wm:-Aqua}"

	# hardware
	cpu="$(sysctl -n machdep.cpu.brand_string)"
	ram="$(sysctl -n hw.memsize)"

	# editor, remove the file path
	[ "$EDITOR" ] && EDITOR="${EDITOR##*/}"


	case $TERM_PROGRAM in
		"Terminal.app" | "Apple_Terminal") term="Apple Terminal";;
		"iTerm.app")    term="iTerm2";;
		*)              term="${TERM_PROGRAM%.app}";;
	esac

	print
else
	echo "Unsupported OS; please add support on https://github.com/unixporn/trup"
fi
