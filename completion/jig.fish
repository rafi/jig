status is-interactive; or exit 0

function __fish_jig_complete_projects
	jig list
end

function __fish_jig_complete_sessions
	tmux ls -F '#S'
end

set -l jig_commands start stop print list edit new switch version

complete -f -c jig -n "not __fish_seen_subcommand_from $jig_commands" -a "$jig_commands"
complete -f -c jig -n "__fish_seen_subcommand_from start stop list edit new; and not __fish_seen_subcommand_from (__fish_jig_complete_projects)" -a "(__fish_jig_complete_projects)"
complete -f -c jig -n "__fish_seen_subcommand_from print switch; and not __fish_seen_subcommand_from (__fish_jig_complete_sessions)" -a "(__fish_jig_complete_sessions)"
