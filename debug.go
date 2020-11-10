package main

func addDebugHandlers(router *Mux) {
	router.Command("/debug/dmesg", "kernel logs",
		"dmesg")
	router.Command("/debug/processes", "process list",
		"ps", "-eo", "user,pid,ppid,c,stime,tty,%cpu,%mem,vsz,rsz,cmd")
	router.Command("/debug/logs", "logs",
		"journalctl", "--reverse", "-b", "--no-pager", "-n", "50")
	router.Command("/debug/meminfo", "memory data",
		"/proc/meminfo")
}
