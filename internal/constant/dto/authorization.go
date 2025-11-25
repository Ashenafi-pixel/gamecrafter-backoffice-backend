package dto

import (
	"net/http"

	"github.com/google/uuid"
)

type Permissions struct {
	ID          uuid.UUID `gorm:"primary_key" json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type Role struct {
	ID          uuid.UUID     `gorm:"primary_key" json:"id"`
	Name        string        `json:"name"`
	Permissions []Permissions `json:"permissions,omitempty"`
}

type PermissionWithValue struct {
	PermissionID uuid.UUID `json:"permission_id"`
	Value        *float64  `json:"value,omitempty"` // NULL = unlimited, value = funding limit
}

type CreateRoleReq struct {
	Name        string                `json:"name"`
	Permissions []PermissionWithValue `json:"permissions"` // Updated to include values
}

type UserRole struct {
	UserID uuid.UUID `json:"user_id"`
	RoleID uuid.UUID `json:"role_id"`
}
type UserRolesRes struct {
	UserID uuid.UUID `json:"user_id"`
	Roles  []Role    `json:"roles"`
}

type PermissionsToRoute struct {
	ID          uuid.UUID `json:"id"`
	EndPoint    string    `json:"endpoint"`
	Name        string    `json:"name"`
	Method      string    `json:"method"`
	Description string    `json:"description"`
}
type UpdatePermissionToRoleReq struct {
	RoleID      uuid.UUID             `json:"role_id"`
	Permissions []PermissionWithValue `json:"permissions"` // Updated to include values
}
type UpdatePermissionToRoleRes struct {
	Message string `json:"message"`
	Role    Role   `json:"role"`
}
type GetPermissionReq struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}
type GetPermissionData struct {
	Permissions Permissions `json:"permission"`
	Roles       []Role      `json:"roles"`
}

type GetPermissionRes struct {
	Message string            `json:"message"`
	Data    GetPermissionData `json:"data"`
}

type RolePermissions struct {
	RoleID      uuid.UUID   `json:"role_id"`
	Permissions []uuid.UUID `json:"permissions"`
}

type AssignPermissionToRoleData struct {
	ID           uuid.UUID `json:"id"`
	RoleID       uuid.UUID `json:"role_id"`
	PermissionID uuid.UUID `json:"permission_id"`
	Value        *float64  `json:"value,omitempty"` // NULL = unlimited, value = funding limit
}

type AssignPermissionToRoleRes struct {
	Message string                     `json:"message"`
	Data    AssignPermissionToRoleData `json:"data"`
}

type GetRoleReq struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type RemoveRoleReq struct {
	RoleID uuid.UUID `json:"role_id"`
}

type AssignRoleToUserReq struct {
	RoleID uuid.UUID `json:"role_id"`
	UserID uuid.UUID `json:"user_id"`
}

type AssignRoleToUserRes struct {
	UserID uuid.UUID `json:"user_id"`
	Roles  []Role    `json:"roles"`
}

var PermissionsList = map[string]PermissionsToRoute{
	"block user account":               {EndPoint: "/api/admin/user/block", Method: http.MethodPost, Name: "block user account", Description: "allow user to block player's account"},
	"add ip filter":                    {EndPoint: "/api/admin/ipfilters", Method: http.MethodPost, Name: "add ip filter", Description: "allow user to add ip filter"},
	"remove ip filter":                 {EndPoint: "/api/admin/ipfilters", Method: http.MethodDelete, Name: "remove ip filter", Description: "allow user to remove ip filter"},
	"get fund logs":                    {EndPoint: "/api/admin/balance/log/funds", Method: http.MethodGet, Name: "get fund logs", Description: "allow user to fetch manual fund logs"},
	"get balance logs":                 {EndPoint: "/api/admin/balance/logs", Method: http.MethodGet, Name: "get balance logs", Description: "allow user to fetch balance logs"},
	"manual funding":                   {EndPoint: "/api/admin/players/funding", Method: http.MethodPost, Name: "add or remove fund manually", Description: "allow user to add or remove fund manually"},
	"get departments":                  {EndPoint: "/api/admin/departments", Method: http.MethodGet, Name: "get departments", Description: "allow user to fetch list of departments"},
	"create departments":               {EndPoint: "/api/admin/departments", Method: http.MethodPost, Name: "create departments", Description: "allow user to create new department"},
	"update department":                {EndPoint: "/api/admin/departments", Method: http.MethodPatch, Name: "update department", Description: "allow user to update department"},
	"assign userto depertment":         {EndPoint: "/api/admin/departments/assign", Method: http.MethodPost, Name: "assign userto depertment", Description: "allow user to assign user to department"},
	"get failed rounds":                {EndPoint: "/api/admin/failed/rounds", Method: http.MethodGet, Name: "get failed rounds", Description: "allow user to fetch failed rounds"},
	"manual refund failed rounds":      {EndPoint: "/api/admin/failed/rounds", Method: http.MethodPost, Name: "manual refund failed rounds", Description: "allow user to refund failed round players"},
	"get bet history":                  {EndPoint: "/api/admin/failed/rounds", Method: http.MethodGet, Name: "get bet history", Description: "allow user to fetch bet history"},
	"get ip filters":                   {EndPoint: "/api/admin/ipfilters", Method: http.MethodGet, Name: "get ip filters", Description: "allow user to fetch ip filters"},
	"get financial metrics":            {EndPoint: "/api/admin/metrics/financial", Method: http.MethodGet, Name: "get financial metrics", Description: "allow user to get financial metrics"},
	"get game metrics":                 {EndPoint: "/api/admin/metrics/game", Method: http.MethodGet, Name: "get game metrics", Description: "allow user to get game metrics"},
	"get blocked account":              {EndPoint: "/api/admin/users/block/accounts", Method: http.MethodPost, Name: "get blocked account", Description: "allow user to get blocked account"},
	"reset user account password":      {EndPoint: "/api/admin/users", Method: http.MethodPost, Name: "reset user account password", Description: "allow user to reset user account password"},
	"update user profile":              {EndPoint: "/api/admin/users", Method: http.MethodPatch, Name: "update user profile", Description: "allow user to update user profile"},
	"add operational group type":       {EndPoint: "/api/admin/operationalgrouptype", Method: http.MethodPost, Name: "add operational group type", Description: "allow to add operational group types"},
	"get operational group type":       {EndPoint: "/api/admin/operationalgrouptype/:groupID", Method: http.MethodGet, Name: "get operational group type", Description: "allow to get operational group type"},
	"add operational group":            {EndPoint: "/api/admin/operationalgroup", Method: http.MethodPost, Name: "add operational group", Description: "allow to add operational group types"},
	"get operational group":            {EndPoint: "/api/admin/operationalgroup", Method: http.MethodGet, Name: "get operational group", Description: "allow to get operational group type"},
	"get players":                      {EndPoint: "/api/admin/users", Method: http.MethodPost, Name: "get players", Description: "allow to get players "},
	"get referral multiplier":          {EndPoint: "/api/admin/referrals", Method: http.MethodGet, Name: "get referral multiplier", Description: "allow user to get the current referral multiplier "},
	"update referral multiplier":       {EndPoint: "/api/admin/referrals", Method: http.MethodPost, Name: "update referral multiplier", Description: "allow user to update the current referral multiplier "},
	"add point to users":               {EndPoint: "/api/admin/referrals/users", Method: http.MethodPost, Name: "add point to users", Description: "allow user to add points to players"},
	"get point to users":               {EndPoint: "/api/admin/referrals/users", Method: http.MethodGet, Name: "get point to users", Description: "allow user to get admin funded points"},
	"create league":                    {EndPoint: "/api/admin/football/leagues", Method: http.MethodPost, Name: "create league", Description: "allow user to create league"},
	"get leagues":                      {EndPoint: "/api/admin/football/leagues", Method: http.MethodGet, Name: "get leagues", Description: "allow user to get leagues"},
	"get clubs":                        {EndPoint: "/api/admin/football/clubs", Method: http.MethodGet, Name: "get clubs", Description: "allow user to get clubs"},
	"create clubs":                     {EndPoint: "/api/admin/football/clubs", Method: http.MethodPost, Name: "create clubs", Description: "allow user to create clubs"},
	"update football match multiplier": {EndPoint: "/api/admin/football/configs", Method: http.MethodPut, Name: "update football match multiplier", Description: "allow user to update football match multiplier"},
	"get football match multiplier":    {EndPoint: "/api/admin/football/configs", Method: http.MethodGet, Name: "get football match multiplier", Description: "allow user to get football match multiplier"},
	"create football match round":      {EndPoint: "/api/admin/football/rounds", Method: http.MethodPost, Name: "create football match round", Description: "allow user to create football match round"},
	"get football match round":         {EndPoint: "/api/admin/football/rounds", Method: http.MethodGet, Name: "get football match round", Description: "allow user to get football match round"},
	"create football matches":          {EndPoint: "/api/admin/football/matches", Method: http.MethodPost, Name: "create football matches", Description: "allow user to create football matches"},
	"get football match round matches": {EndPoint: "/api/admin/football/rounds/mathces", Method: http.MethodPost, Name: "get football match round matches", Description: "allow user to get football match round matches"},
	"close football matche":            {EndPoint: "/api/admin/football/rounds/mathces", Method: http.MethodPatch, Name: "close football matche", Description: "allow user to close football matche"},
	"udpate football round price":      {EndPoint: "/api/admin/football/rounds/prices", Method: http.MethodPost, Name: "udpate football round price", Description: "allow user to udpate football round price"},
	"update cryptokings config":        {EndPoint: "/api/admin/cryptokings/configs", Method: http.MethodPost, Name: "update cryptokings config", Description: "allow admin to update crypto kings config "},
	"get football round price":         {EndPoint: "/api/admin/football/rounds/prices", Method: http.MethodGet, Name: "get football round price", Description: "allow user to get football round price"},
	"update airtime price":             {EndPoint: "/api/admin/airtime/price", Method: http.MethodPut, Name: "update airtime price", Description: "allow user to update airtime price"},
	"refresh airtime utilities":        {EndPoint: "/api/admin/airtime/refresh", Method: http.MethodGet, Name: "refresh airtime utilities", Description: "allow user to fresh database to get updated utilities"},
	"get airtime utilities":            {EndPoint: "/api/admin/airtime/utilities", Method: http.MethodGet, Name: "get airtime utilities", Description: "allow user to get airtime utilities"},
	"update airtime utility status":    {EndPoint: "/api/admin/airtime", Method: http.MethodPut, Name: "update airtime utility status", Description: "allow user to update airtime status"},
	"get airtime transactions":         {EndPoint: "/api/admin/airtime/transactions", Method: http.MethodGet, Name: "get airtime transactions", Description: "allow user to get airtime transactions"},
	"Get Audit Logs":                   {EndPoint: "/api/admin/log/audit", Method: http.MethodPost, Name: "Get Audit Logs", Description: "allow user to list Audit Logs"},
	"Get Available Logs Module":        {EndPoint: "/api/admin/logs/modules", Method: http.MethodGet, Name: "Get Available Logs Module", Description: "allow user to Get Available Logs Module"},
	"register players":                 {EndPoint: "/api/admin/players/register", Method: http.MethodPost, Name: "register players", Description: "allow user to register players"},
	"get games":                        {EndPoint: "/api/admin/games", Method: http.MethodGet, Name: "get games", Description: "allow user to get games"},
	"update games":                     {EndPoint: "/api/admin/games", Method: http.MethodPut, Name: "update games", Description: "allow user to update games"},
	"disable games":                    {EndPoint: "/api/admin/games/disable", Method: http.MethodPost, Name: "disable games", Description: "allow user to disable games"},
	"get admins":                       {EndPoint: "/api/admins", Method: http.MethodGet, Name: "get admins", Description: "allow user to list admins"},
	"update airtime amount":            {EndPoint: "/api/admin/airtime/amount", Method: http.MethodPut, Name: "update airtime amount", Description: "allow user to update airtime amount"},
	"get airtime utilities stats":      {EndPoint: "/api/admin/airtime/stats", Method: http.MethodGet, Name: "get airtime utilities stats", Description: "allow user to get airtime utilities stats"},
	"create company":                   {EndPoint: "/api/admin/companies", Method: http.MethodPost, Name: "create company", Description: "allow user to create company"},
	"get companies":                    {EndPoint: "/api/admin/companies", Method: http.MethodGet, Name: "get companies", Description: "allow user to get companies"},
	"update company":                   {EndPoint: "/api/admin/companies", Method: http.MethodPatch, Name: "update company", Description: "allow user to update company"},
	"delete company":                   {EndPoint: "/api/admin/companies", Method: http.MethodDelete, Name: "delete company", Description: "allow user to delete company"},
	"add ip to company":                {EndPoint: "/api/admin/companies/:id/add-ip", Method: http.MethodPatch, Name: "add ip to company", Description: "allow user to add ip to company"},
	"get company":                      {EndPoint: "/api/admin/companies/:id", Method: http.MethodGet, Name: "get company", Description: "allow user to get company"},
	"get available games":              {EndPoint: "/api/admin/games/available", Method: http.MethodGet, Name: "get available games", Description: "allow user to get available games"},
	"delete games":                     {EndPoint: "/api/admin/games", Method: http.MethodDelete, Name: "delete games", Description: "allow user to delete games"},
	"add games":                        {EndPoint: "/api/admin/games", Method: http.MethodPost, Name: "add games", Description: "allow user to add games"},
	"update game status":               {EndPoint: "/api/admin/games/status", Method: http.MethodPut, Name: "update game status", Description: "allow admin to update game status"},
	"get daily report":                 {EndPoint: "/api/admin/report/daily", Method: http.MethodGet, Name: "get daily report", Description: "allow user to get daily report"},
	"get duplicate ip accounts report": {EndPoint: "/api/admin/report/duplicate-ip-accounts", Method: http.MethodGet, Name: "get duplicate ip accounts report", Description: "allow user to get duplicate IP accounts report"},
	"get big winners report":           {EndPoint: "/api/admin/report/big-winners", Method: http.MethodGet, Name: "get big winners report", Description: "allow user to get big winners report"},
	"get player metrics report":        {EndPoint: "/api/admin/report/player-metrics", Method: http.MethodGet, Name: "get player metrics report", Description: "allow user to get player metrics report"},
	"get player transactions report":   {EndPoint: "/api/admin/report/player-metrics/:player_id/transactions", Method: http.MethodGet, Name: "get player transactions report", Description: "allow user to get player transactions for drill-down"},
	"get country report":                {EndPoint: "/api/admin/report/country", Method: http.MethodGet, Name: "get country report", Description: "allow user to get country report"},
	"get game performance report":        {EndPoint: "/api/admin/report/game-performance", Method: http.MethodGet, Name: "get game performance report", Description: "allow user to get game performance report"},
	"get game players report":            {EndPoint: "/api/admin/report/game-performance/:game_id/players", Method: http.MethodGet, Name: "get game players report", Description: "allow user to get players who played a specific game"},
	"get provider performance report":     {EndPoint: "/api/admin/report/provider-performance", Method: http.MethodGet, Name: "get provider performance report", Description: "allow user to get provider performance report"},
	"create mysteries":                 {EndPoint: "/api/admin/spinningwheels/mysteries", Method: http.MethodPost, Name: "create mysteries", Description: "allow user to create spinning wheel mysteries"},
	"get mysteries":                    {EndPoint: "/api/admin/spinningwheels/mysteries", Method: http.MethodGet, Name: "get mysteries", Description: "allow user to get spinning wheel mysteries"},
	"delete mysteries":                 {EndPoint: "/api/admin/spinningwheels/mysteries", Method: http.MethodDelete, Name: "delete mysteries", Description: "allow user to delete spinning wheel mysteries"},
	"update mysteries":                 {EndPoint: "/api/admin/spinningwheels/mysteries", Method: http.MethodPut, Name: "update mysteries", Description: "allow user to update spinning wheel mysteries"},
	"create spinning wheel config":     {EndPoint: "/api/admin/spinningwheels/config", Method: http.MethodPost, Name: "create spinning wheel config", Description: "allow user to create spinning wheel config"},
	"get spinning wheel configs":       {EndPoint: "/api/admin/spinningwheels/configs", Method: http.MethodGet, Name: "get spinning wheel configs", Description: "allow user to get spinning wheel configs"},
	"update spinning wheel config":     {EndPoint: "/api/admin/spinningwheels/config", Method: http.MethodPut, Name: "update spinning wheel config", Description: "allow user to update spinning wheel config"},
	"delete spinning wheel config":     {EndPoint: "/api/admin/spinningwheels/config", Method: http.MethodDelete, Name: "delete spinning wheel config", Description: "allow user to delete spinning wheel config"},
	"update bet icon":                  {EndPoint: "/api/admin/bets/icons", Method: http.MethodPost, Name: "update bet icon", Description: "allow user to update bet icon"},
	"get scratch card configs":         {EndPoint: "/api/admin/scratchcards/configs", Method: http.MethodGet, Name: "get scratch card configs", Description: "allow user to get scratch card configs"},
	"update scratch card configs":      {EndPoint: "/api/admin/scratchcards/configs", Method: http.MethodPut, Name: "update scratch card configs", Description: "allow user to update scratch card configs"},
	"create bet level":                 {EndPoint: "/api/admin/bet/levels", Method: http.MethodPost, Name: "create bet level", Description: "allow user to create bet level"},
	"get bet levels":                   {EndPoint: "/api/admin/bet/levels", Method: http.MethodGet, Name: "get bet levels", Description: "allow user to get bet levels"},
	"create level requirements":        {EndPoint: "/api/admin/bet/levels/requirements", Method: http.MethodPost, Name: "create level requirements", Description: "allow user to create level requirements"},
	"update level requirements":        {EndPoint: "/api/admin/bet/levels/requirements", Method: http.MethodPatch, Name: "update level requirements", Description: "allow user to update level requirements"},
	"update signup bonus":              {EndPoint: "/api/admin/signup/bonus", Method: http.MethodPut, Name: "update signup bonus", Description: "allow user to update signup bonus"},
	"get signup bonus":                 {EndPoint: "/api/admin/signup/bonus", Method: http.MethodGet, Name: "get signup bonus", Description: "allow user to get signup bonus"},
	"Create tournament":                {EndPoint: "/api/admins/tournaments", Method: http.MethodPost, Name: "Create tournament", Description: "allow user to create tournament"},
	"Create Loot Box":                  {EndPoint: "/api/admin/lootboxes", Method: http.MethodPost, Name: "Create Loot Box", Description: "allow user to create loot box"},
	"Get Loot Box":                     {EndPoint: "/api/admin/lootboxes", Method: http.MethodGet, Name: "Get Loot Box", Description: "allow user to get loot box"},
	"Update Loot Box":                  {EndPoint: "/api/admin/lootboxes", Method: http.MethodPut, Name: "Update Loot Box", Description: "allow user to update loot box"},
	"Delete Loot Box":                  {EndPoint: "/api/admin/lootboxes/:id", Method: http.MethodDelete, Name: "Delete Loot Box", Description: "allow user to delete loot box"},
	"get adds services":                {EndPoint: "/api/admin/adds/services", Method: http.MethodGet, Name: "get adds services", Description: "allow user to get adds services"},
	"create adds service":              {EndPoint: "/api/admin/adds/services", Method: http.MethodPost, Name: "create adds service", Description: "allow user to create adds service"},
	"banner read":                      {EndPoint: "/api/admin/banners", Method: http.MethodGet, Name: "banner read", Description: "allow user to read banner information"},
	"banner display":                   {EndPoint: "/api/admin/banners/display", Method: http.MethodGet, Name: "banner display", Description: "allow user to display banner information"},
	"banner update":                    {EndPoint: "/api/admin/banners/:id", Method: http.MethodPatch, Name: "banner update", Description: "allow user to update banner information"},
	"banner create":                    {EndPoint: "/api/admin/banners", Method: http.MethodPost, Name: "banner create", Description: "allow user to create banner information"},
	"banner delete":                    {EndPoint: "/api/admin/banners/:id", Method: http.MethodDelete, Name: "banner delete", Description: "allow user to delete banner information"},
	"banner image upload":              {EndPoint: "/api/admin/banners/upload", Method: http.MethodPost, Name: "banner image upload", Description: "allow user to upload banner images"},
	"Create Lottery Service":           {EndPoint: "/api/admin/lottery/service", Method: http.MethodPost, Name: "Create Lottery Service", Description: "allow user to create lottery service"},
	"Create Lottery":                   {EndPoint: "/admin/lottery/request", Method: http.MethodPost, Name: "Create Lottery", Description: "allow user to create lottery"},
	"Create Agent Provider":            {EndPoint: "/api/admin/agent/providers", Method: http.MethodPost, Name: "Create Agent Provider", Description: "allow user to create agent provider"},
	"Get Agent Referrals":              {EndPoint: "/api/admin/agent/referrals", Method: http.MethodGet, Name: "Get Agent Referrals", Description: "allow user to get agent referrals"},
	"Get Agent Referral Stats":         {EndPoint: "/api/admin/agent/stats", Method: http.MethodGet, Name: "Get Agent Referral Stats", Description: "allow user to get agent referral stats"},

	// role related
	"get permissions":         {EndPoint: "/api/admin/permissions", Method: http.MethodGet, Name: "get permissions", Description: "allow user to get list of permissions"},
	"create role":             {EndPoint: "/api/admin/roles", Method: http.MethodPost, Name: "create role", Description: "allow user to create role"},
	"assign role":             {EndPoint: "/api/admin/users/roles", Method: http.MethodPost, Name: "assign role", Description: "allow user to assign role to other user"},
	"revoke role":             {EndPoint: "/api/admin/roles", Method: http.MethodDelete, Name: "revoke role", Description: "allow user to revoke other user role"},
	"update role permissions": {EndPoint: "/api/admin/roles", Method: http.MethodPatch, Name: "update role permissions", Description: "allow user to update role permissions"},
	"remove role":             {EndPoint: "/api/admin/roles", Method: http.MethodDelete, Name: "remove role", Description: "allow user to remove role"},
	"get roles":               {EndPoint: "/api/admin/roles", Method: http.MethodGet, Name: "get roles", Description: "allow user to get roles"},
	"get role users":          {EndPoint: "/api/admin/roles/:id", Method: http.MethodGet, Name: "get role users", Description: "allow user to get role users"},
	"get user roles":          {EndPoint: "/api/admin/users/:id/roles/", Method: http.MethodGet, Name: "get user roles", Description: "allow user to get user roles"},
	"super":                   {EndPoint: "*", Method: "*", Name: "super", Description: "supper user has all permissions on the system"},
}
