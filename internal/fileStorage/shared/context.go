package shared

type contextKey string

const IsLoggedInContextKey = contextKey("isLoggedIn")
const IsAdminContextKey = contextKey("isAdmin")
const UserModelContextKey = contextKey("userStruct")
const IsSubscribedContextKey = contextKey("isSubscribed")
