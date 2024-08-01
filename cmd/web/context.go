package main

type contextKey string

const isLoggedInContextKey = contextKey("isLoggenIn")

// const userIdContextKey = contextKey("userId")
const userModelContextKey = contextKey("userStruct")
const isSubscribedContextKey = contextKey("isSubscribed")
