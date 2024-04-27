package main

type contextKey string

const isLoggedInContextKey = contextKey("isAuthenticated")
const userIdContextKey = contextKey("userId")

