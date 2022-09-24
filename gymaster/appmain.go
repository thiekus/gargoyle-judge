package main

/* GargoyleJudge - Simple Judgement System for Competitive Programming
 * Copyright (C) Thiekus 2019
 * Visit www.khayalan.id for updates
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

import (
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	"github.com/thiekus/gargoyle-judge/internal/gylib"
)

const appVersion = "0.8r209"

var appOSName string
var appConfig ConfigData

var appServer http.Server
var appServerVer = fmt.Sprintf("ThkGargoyleWS %s", appVersion)
var appOnShutdown = false
var appOnRestart = false
var appUsers UserController
var appSlaves SlaveManager
var appContestAccess ContestAccessController
var appLangPrograms LanguageProgramController
var appScoreboard ScoreboardController
var appNotifications NotificationController

func FixRootPath(oldPath string) string {
	if oldPath == "/" {
		oldPath = ""
	}
	if (appConfig.RootSubPath == "") || (appConfig.RootSubPath == "/") {
		return oldPath
	} else {
		return appConfig.RootSubPath + oldPath
	}
}

func GetAppRootPrefix() string {
	if (appConfig.RootSubPath == "") || (appConfig.RootSubPath == "/") {
		return ""
	} else {
		return appConfig.RootSubPath
	}
}

func GetAppUrl(r *http.Request) string {
	if (appConfig.RootSubPath == "") || (appConfig.RootSubPath == "/") {
		return gylib.GetBaseUrl(r)
	} else {
		return gylib.GetBaseUrl(r) + appConfig.RootSubPath
	}
}

func GetAppUrlWithSlash(r *http.Request) string {
	if (appConfig.RootSubPath == "") || (appConfig.RootSubPath == "/") {
		return gylib.GetBaseUrlWithSlash(r)
	} else {
		return gylib.GetBaseUrl(r) + appConfig.RootSubPath + "/"
	}
}

// Endpoint to perform application shutdown from http request.
// Needs authentication to admin user.
func shutdownEndpoint(w http.ResponseWriter, r *http.Request) {
	// Check authentication before do shutdown
	ui := appUsers.GetLoggedUserInfo(r)
	if (ui != nil) && (ui.IsAdmin()) {
		log := gylib.GetStdLog()
		log.Print("Requesting shutdown...")
		appOnShutdown = true
		go func() {
			time.Sleep(5000 * time.Millisecond)
			log.Print("Shutting down...")
			if err := appServer.Shutdown(context.Background()); err != nil {
				log.Printf("Shutdown error: %s", err)
			}
		}()
		fmt.Fprintf(w, "Goodbye! Server will be shutdown in 5 seconds...")
	} else {
		http.Redirect(w, r, "login", 302)
	}
}

func aboutEndpoint(w http.ResponseWriter, r *http.Request) {
	CompileSinglePage(w, r, "about.html", nil)
}

// Main middleware, invoking some tweaks
func appMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := gylib.GetStdLog()
		w.Header().Add("Server", appServerVer)
		w.Header().Set("Access-Control-Allow-Origin", gylib.GetBaseUrl(r))
		if appOnShutdown {
			log.Printf("Request for client %s to %s rejected on shutdown", r.RemoteAddr, r.URL.Path)
			http.Error(w, "Internal server error: server on shutdown", http.StatusInternalServerError)
			return
		}
		// Print access log, if not ajax
		uid := appUsers.GetLoggedUserId(r)
		path := r.URL.Path
		query := r.URL.RawQuery
		if query != "" {
			path += "?" + query
		}
		w.Header().Set("Location", GetAppUrl(r)+path)
		// All ajax endpoints are json-return
		if strings.HasPrefix(r.URL.Path, FixRootPath("/ajax")) {
			w.Header().Add("Content-Type", "application/json; charset=utf-8")
		} else {
			log.Printf("Client %s uid:%d accessing %s", r.RemoteAddr, uid, path)
		}
		// Check user is login?
		user := appUsers.GetLoggedUserInfo(r)
		// Restrict dashboard from stranger
		if strings.HasPrefix(r.URL.Path, FixRootPath("/dashboard")) {
			urlBase64 := base64.StdEncoding.EncodeToString([]byte(path))
			if user == nil {
				appUsers.AddFlashMessage(w, r, "Please login first!", FlashError)
				http.Redirect(w, r, GetAppUrl(r)+"/login?target="+urlBase64, 302)
				return
			} else {
				// User has been banned
				if user.Banned {
					log.Errorf("uid:%d cannot access dashboard because have been banned!", user.Id)
					appUsers.UserLogoutFromWebsite(w, r)
					appUsers.AddFlashMessage(w, r, "You have been banned!", FlashError)
					http.Redirect(w, r, GetAppUrl(r)+"/login?target="+urlBase64, 302)
					return
				}
				// User inactive
				if !user.Active {
					log.Errorf("uid:%d cannot access dashboard because account was inactive!", user.Id)
					appUsers.UserLogoutFromWebsite(w, r)
					appUsers.AddFlashMessage(w, r, "You account is not activated or deactivated by admin!", FlashError)
					http.Redirect(w, r, GetAppUrl(r)+"/login?target="+urlBase64, 302)
					return
				}
				// Further page access permission
				access, err := GetPageAccessPermission()
				if err != nil {
					http.Error(w, "500 Internal Server Error: "+err.Error(), http.StatusInternalServerError)
					return
				}
				for _, v := range access {
					if strings.HasPrefix(r.URL.Path, FixRootPath(v.Prefix)) {
						if !IsPageAccessHasPermission(v, user.Roles) {
							log.Errorf("uid:%d cannot access dashboard because insufficient privileges! Roles: %v", user.Id, user.Roles)
							http.Error(w, "403 Forbidden: insufficient privileges", http.StatusForbidden)
							return
						}
					}
				}
			}
		} else if (strings.HasPrefix(r.URL.Path, FixRootPath("/login")) || (r.URL.Path == FixRootPath("/"))) && (user != nil) {
			http.Redirect(w, r, GetAppUrl(r)+"/dashboard", 302)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func prepareConfig() {
	// Get configuration data
	appConfig = getConfigData()
}

func prepareDatabase() {
	log := gylib.GetStdLog()
	log.Print("Testing database connection...")
	log.Printf("DB Driver: %s", appConfig.DbDriver)
	db, err := OpenDatabaseEx(appConfig.DbDriver, false)
	if err != nil {
		log.Error(err)
	} else {
		defer db.Close()
	}
	log.Print("Ping test into selected database...")
	if err = db.Ping(); err != nil {
		log.Errorf("Ping error: %s", err.Error())
	} else {
		log.Print("DB Connection OK!")
	}
	if !appConfig.HasFirstSetup {
		log.Warn("You seems never do First Setup. Please do that before someone broke your mind!")
	}
}

func prepareControllers() {
	appUsers = MakeUserController()
	appSlaves = MakeSlaveManager()
	appContestAccess = MakeContestAccessController()
	appLangPrograms = MakeLanguageProgramController()
	appScoreboard = MakeScoreboardController()
	appNotifications = MakeNotificationController()
}

func prepareHttpEndpoints() {
	log := gylib.GetStdLog()
	// Define our webservice routing
	r := mux.NewRouter()
	r.Use(appMiddleware)

	// Main webservice endpoints, see frontpage.go
	r.HandleFunc(FixRootPath("/"), homeGetEndpoint).Methods("GET")
	if GetAppRootPrefix() != "" {
		// Fix for subdirectory level path
		r.HandleFunc(GetAppRootPrefix(), homeGetEndpoint).Methods("GET")
	}
	// see userauth.go
	r.HandleFunc(FixRootPath("/login"), loginGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/login"), loginPostEndpoint).Methods("POST")
	r.HandleFunc(FixRootPath("/logout"), logoutGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/forgotPass"), forgotPassGetEndpoint).Methods("GET")
	// see dashboard_basic.go
	r.HandleFunc(FixRootPath("/dashboard"), dashboardHomeGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/dashboard/notifications"), dashboardNotificationsEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/dashboard/scoreboard"), dashboardScoreboardsGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/dashboard/scoreboard/{id}"), dashboardViewScoreboardGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/dashboard/profile"), dashboardProfileGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/dashboard/profile"), dashboardProfilePostEndpoint).Methods("POST")
	r.HandleFunc(FixRootPath("/dashboard/settings"), dashboardSettingsGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/dashboard/settings"), dashboardSettingsPostEndpoint).Methods("POST")
	// see dashboard_contestant.go
	r.HandleFunc(FixRootPath("/dashboard/contests"), dashboardContestsGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/dashboard/problemSet/{id}"), dashboardProblemSetGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/dashboard/problem/{id}"), dashboardProblemGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/dashboard/problem"), dashboardProblemPostEndpoint).Methods("POST")
	r.HandleFunc(FixRootPath("/dashboard/userSubmissions"), dashboardUserSubmissionsGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/dashboard/userViewSubmission/{id}"), dashboardUserViewSubmissionGetEndpoint).Methods("GET")
	// see dashboard_jury.go
	r.HandleFunc(FixRootPath("/dashboard/manageContests"), dashboardManageContestsGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/dashboard/contestAdd"), dashboardContestAddGetEndpoint).Methods("GET")
	// see dashboard_admin.go
	r.HandleFunc(FixRootPath("/dashboard/manageUsers"), dashboardManageUsersGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/dashboard/userAdd"), dashboardUserAddGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/dashboard/userAdd"), dashboardUserAddPostEndpoint).Methods("POST")
	r.HandleFunc(FixRootPath("/dashboard/userEdit/{id}"), dashboardUserEditGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/dashboard/userEdit"), dashboardUserEditPostEndpoint).Methods("POST")
	r.HandleFunc(FixRootPath("/dashboard/userDelete/{id}"), dashboardUserDeleteGetEndpoint).Methods("GET")
	// see ajax_users.go
	r.HandleFunc(FixRootPath("/ajax/getNotifications"), ajaxGetNotifications).Methods("GET")
	r.HandleFunc(FixRootPath("/ajax/readAllNotifications"), ajaxReadAllNotifications).Methods("GET")
	// see firstsetup.go
	r.HandleFunc(FixRootPath("/gysetup"), firstSetupGetEndpoint).Methods("GET")
	r.HandleFunc(FixRootPath("/gysetup"), firstSetupPostEndpoint).Methods("POST")

	// Resources endpoints, assets and website favicon
	if appConfig.AssetsCaching {
		setAssetsWithCaching(r)
	} else {
		r.PathPrefix(FixRootPath("/assets/")).
			Handler(http.StripPrefix(FixRootPath("/assets/"), http.FileServer(http.Dir(gylib.ConcatByProgramLibDir("./assets")))))
	}
	// Handle favicon
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadFile(gylib.ConcatByProgramLibDir("./favicon.ico"))
		if err == nil {
			w.Header().Set("Cache-Control", "public, max-age=3600")
			w.Header().Set("Content-Type", "image/x-icon")
			w.Header().Set("Content-Length", strconv.Itoa(len(b)))
			w.Write(b)
		} else {
			http.Error(w, "500 Internal Server Error", 500)
		}
	})
	// Private files
	filesDir := gylib.ConcatByProgramLibDir("./files")
	if !gylib.IsDirectoryExists(filesDir) {
		if err := os.Mkdir(filesDir, os.ModePerm); err != nil {
			panic(err)
		}
	}
	r.PathPrefix(FixRootPath("/files/")).Handler(http.StripPrefix(FixRootPath("/files/"), http.FileServer(http.Dir(filesDir))))
	// Avatar images
	r.HandleFunc(FixRootPath("/avatar/{avatarInfo}"), avatarGetEndpoint).Methods("GET")

	// About contents endpoints
	r.HandleFunc(FixRootPath("/about"), aboutEndpoint).Methods("GET")
	// Shutdown endpoints
	r.HandleFunc(FixRootPath("/shutdown"), shutdownEndpoint).Methods("GET")

	log.Print("Routers have been initialized...")

	// Establish our server
	h := http.Handler(r)
	if appConfig.CompressOnFly {
		gzContentType := []string{
			"text/html",
			"application/javascript",
			"text/javascript",
			"application/ecmascript",
			"text/ecmascript",
			"text/css",
			"text/json",
			"application/json",
		}
		gzContentOpt := gziphandler.ContentTypes(gzContentType)
		gh, err := gziphandler.GzipHandlerWithOpts(gzContentOpt,
			gziphandler.CompressionLevel(gzip.DefaultCompression), gziphandler.MinSize(gziphandler.DefaultMinSize))
		if err != nil {
			log.Errorf("Error when setting gzip: %s", err.Error())
			panic(err)
		}
		h = gh(r)
	}
	appServer = http.Server{
		Addr:    fmt.Sprintf(":%d", appConfig.ListeningPort),
		Handler: h,
	}
	if appConfig.UseTLS {
		log.Print("Server will use TLS")
		if err := appServer.ListenAndServeTLS(appConfig.CrtFile, appConfig.KeyFile); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	} else {
		log.Print("Server will use plain HTTP")
		if err := appServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}

}

// Main Frontend Entry-point
func main() {
	fmt.Printf("Gargoyle Judgement System v%s (Master Server)\n", appVersion)
	fmt.Println("Copyright (C) Thiekus 2019")
	fmt.Printf("Built using %s\n", runtime.Version())
	if osName, err := gylib.GetOSName(); err != nil {
		panic(err)
	} else {
		appOSName = osName
	}
	fmt.Printf("Running on %s\n\n", appOSName)

	log := gylib.GetStdLog()
	log.Printf("ProgramDir: %s", gylib.GetProgramLibDir())
	log.Printf("WorkDir: %s", gylib.GetWorkDir())
	for {
		log.Print("Initializing master server...")
		// Invalidate maintenance state
		appOnShutdown = false
		appOnRestart = false
		// Prepare now
		prepareConfig()
		prepareDatabase()
		prepareControllers()
		prepareHttpEndpoints()
		if !appOnRestart {
			break
		}
		log.Print("Restarting master server...")
	}
}
