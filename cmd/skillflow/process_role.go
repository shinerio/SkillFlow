package main

type processRole string

const (
	processRoleDaemon     processRole = "daemon"
	processRoleDaemonOnly processRole = "daemon-only"
	processRoleUI         processRole = "ui"

	internalDaemonFlag     = "--internal-daemon"
	internalDaemonOnlyFlag = "--daemon-only"
	internalUIFlag         = "--internal-ui"
)

var activeProcessRole = processRoleDaemon

func determineProcessRole(args []string) (processRole, []string) {
	if len(args) == 0 {
		return processRoleDaemon, args
	}

	filtered := make([]string, 0, len(args))
	role := processRoleDaemon
	for _, arg := range args {
		switch arg {
		case internalUIFlag:
			role = processRoleUI
		case internalDaemonOnlyFlag:
			role = processRoleDaemonOnly
		case internalDaemonFlag:
			role = processRoleDaemon
		default:
			filtered = append(filtered, arg)
		}
	}
	return role, filtered
}
