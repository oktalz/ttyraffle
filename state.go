package main

type state string

const (
	StateWaitProgress state = "WaitingProgress"
	StateRunning      state = "Running"
	StateResetting    state = "Resetting"
	StatePaused       state = "Paused"
	StateAddUser      state = "AddUser"
	StateDeleteUser   state = "DeleteUser"
	StateChooseFile   state = "ChooseFile"
)
