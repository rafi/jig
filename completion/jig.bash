#!/usr/bin/env bash
# shellcheck disable=SC2207

_jig() {
	local cur="${COMP_WORDS[COMP_CWORD]}"
	local cmds='start stop print list edit new switch version'
	local opts=$'--file --detach --debug --inside --help'

	# Commands
	if [ "${#COMP_WORDS[@]}" -eq 2 ]; then
		COMPREPLY=($(compgen -W "$cmds" -- "${cur}"))
	fi
	if [ "${#COMP_WORDS[@]}" -lt 3 ]; then
		return
	fi

	local prev="${COMP_WORDS[COMP_CWORD - 1]}"

	# Projects
	if [ "${#COMP_WORDS[@]}" -eq 3 ]; then
		case ${prev} in
		sta | star | start | sto | stop | l | ls | list | e | ed | edit | n | ne | new)
			COMPREPLY=($(compgen -W "$(jig list)" -- "${cur}"))
			;;
		p | pr | print | sw | swi | switch)
			COMPREPLY=($(compgen -W "$(tmux ls -F '#S')" -- "${cur}"))
			;;
		esac
		return
	fi

	# Flags
	case $prev in
	-w | --windows) return ;;
	start | stop) opts="$opts --windows" ;;
	esac

	# Suggest options that were not specified already
	for word in "${COMP_WORDS[@]}"; do
		case $word in
		-f | --file) opts="${opts/--file/}" ;;
		-d | --detach) opts="${opts/--detach/}" ;;
		-w | --windows) opts="${opts/--windows/}" ;;
		-i | --inside) opts="${opts/--inside/}" ;;
		--debug) opts="${opts/--debug/}" ;;
		--help) opts="${opts/--help/}" ;;
		esac
	done
	COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
}

complete -F _jig jig
