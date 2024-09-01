package main

type contextKey string

const isLoggedInContextKey = contextKey("isLoggedIn")
const isAdminContextKey = contextKey("isAdmin")

// const userIdContextKey = contextKey("userId")
const userModelContextKey = contextKey("userStruct")
