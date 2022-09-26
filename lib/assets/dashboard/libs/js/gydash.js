/* Gargoyle Dashboard script utilities - Copyright Thiekus 2019 */
/* This script is licensed under MPL 2.0 terms */

// Notification last update timestamp
var notificationLastUpdate = 0;
var pingSound = document.getElementById("gyPingAudio");

// Prevent race conditions while sidebar change state
var sidebarChangeLock = false;
var sidebarHovered = false;

// TickTock
var ttRemainTime = 0

// Roll down dashboard message while showing
function animateDashboardMessage() {
    $("#gyDashboardMessage").delay(250).slideDown(500);
}

function appendZero(num) {
    return num > 9 ? num : "0" + num;
}

function getBaseUrl() {
    return window.location.origin;
}

// Restore sidebar like normal while hovered
function maximizeSidebar(animate) {
    if (sidebarChangeLock) {
        setTimeout(function () {
            maximizeSidebar(animate);
        }, 500);
        return;
    }
    sidebarChangeLock = true;
    $(".nav-left-sidebar").removeAttr("minimized");
    $(".nav-left-sidebar").css("width", "");
    $(".dashboard-wrapper").css("margin-left", "");
    $("#gyDashboardMenus > ul.navbar-nav > li.nav-item").css("width", "");
    if (animate) {
        $(".nav-div-label").slideDown(250);
        $(".nav-link-label").slideDown(250, function () {
            sidebarChangeLock = false;
        });
    } else {
        $(".nav-link-label").show();
        $(".nav-div-label").show();
        sidebarChangeLock = false;
    }
}

function minimizeSidebar(animate) {
    var w = $(window).width();
    var h = $(window).height();
    if (h > w) {
        return;
    }
    if (sidebarChangeLock) {
        setTimeout(function () {
            minimizeSidebar(animate)
        }, 500);
        return;
    }
    sidebarChangeLock = true;
    $(".nav-left-sidebar").attr("minimized", "yes");
    $(".nav-left-sidebar").css("width", 70 + "px");
    $(".dashboard-wrapper").css("margin-left", 70 + "px");
    $("#gyDashboardMenus > ul.navbar-nav > li.nav-item").css("width", 42 + "px");
    $(".slimScrollBar").hide();
    $(".nav-div-label").hide();
    if (animate) {
        $(".nav-link-label").slideUp(250, function () {
            sidebarChangeLock = false;
        });
    } else {
        $(".nav-link-label").hide();
        sidebarChangeLock = false;
    }
}

function showDashboardLoading() {
    $(".loading").show();
    $("#gyDashboardContent").addClass("blur-effect");
}

function hideDashboardLoading() {
    $("#gyDashboardContent").removeClass("blur-effect");
    $(".loading").fadeOut(200);
}

function timeConverter(waktu) {
    var hour = ("0" + (Math.floor(waktu / 3600))).slice(-2);
    var min = ("0" + Math.floor((waktu - (hour * 3600)) / 60)).slice(-2);
    var sec = ("0" + Math.floor(waktu - (hour * 3600) - (min * 60))).slice(-2);
    var time = hour + ':' + min + ':' + sec;
    return time;
}

function tickTock() {
    var waktuSisa = timeConverter(ttRemainTime);
    //console.log(waktuSisa);
    $("#tiktok").html(waktuSisa);
    if (ttRemainTime > 0) {
        ttRemainTime--;
        setTimeout(function () {
            tickTock();
        }, 1000);
    } else {
        window.location = getBaseUrl() + "/codeinfest/dashboard?timeOut=yes";
    }
}

function setTickTock(waktu) {
    if (ttRemainTime <= 0) {
        ttRemainTime = waktu;
        tickTock();
    } else {
        ttRemainTime = waktu;
    }
}

function dashboardClock() {
    var nowTime = new Date();
    var timeStr = appendZero(nowTime.getHours()) + ":" +
        appendZero(nowTime.getMinutes()) + ":" +
        appendZero(nowTime.getSeconds());
    $("#gyClock").html(timeStr);
    setTimeout(dashboardClock, 1000);
}

function progressiveDashboardPageLoadSuccess(resp, status, xhr, pushPage) {
    var location = xhr.getResponseHeader('Location');
    if (location.includes(getBaseUrl() + "/codeinfest/login")) {
        window.location = location;
        return;
    }
    var parsedResponse = $.parseHTML(resp, document, true);
    $("#gyDashboardContent").fadeOut(200, function () {
        // Before transplated
        var content = $(parsedResponse).find("#gyDashboardContent").html();
        $("#gyDashboardContent").html(content);
        var menubar = $(parsedResponse).find("#gyDashboardMenus");
        var menulink = $(menubar).find(".nav-link");
        $(menulink).each(function (index, elem) {
            var elemId = $(elem).attr("id");
            $(elem).hasClass("active") ? $("#" + elemId).addClass("active") : $("#" + elemId).removeClass("active");
        });
        $(".os-viewport").scrollTop(0);
        if (pushPage) {
            history.pushState(null, "", location);
        }
        hideDashboardLoading();
        $("#gyDashboardContent").fadeIn(200, function () {
            // After transplated
            try {
                subviewInit();
            } catch (err) {
                alert(err);
            }
            redefineAnchorVisit();
            animateDashboardMessage();
        });
    });
}

function progressiveDashboardPageGet(href, pushPage) {
    try {
        showDashboardLoading();
        $.ajax({
            type: "GET",
            url: href,
            success: function (resp, status, xhr) {
                progressiveDashboardPageLoadSuccess(resp, status, xhr, pushPage);
            },
            error: function (xhr, reason, ex) {
                hideDashboardLoading();
                alert(reason + ": " + ex);
            }
        });
    } catch (err) {
        hideDashboardLoading();
        alert(err);
    }
}

function progressiveDashboardPagePost(action, formData) {
    try {
        showDashboardLoading();
        $.ajax({
            type: "POST",
            url: action,
            data: formData,
            success: function (resp, status, xhr) {
                progressiveDashboardPageLoadSuccess(resp, status, xhr, true);
            },
            error: function (xhr, reason, ex) {
                hideDashboardLoading();
                alert(reason + ": " + ex);
            }
        });
    } catch (err) {
        hideDashboardLoading();
        alert(err);
    }
}

function redefineAnchorVisit() {
    $("a").each(function () {
        var anchor = $(this);
        var href = anchor.attr("href");
        if (typeof href === "undefined") {
            return;
        }
        var isDashboardLink = href.includes(getBaseUrl() + "/codeinfest/dashboard");
        var isQuickLink = $(anchor).attr("gyQuickLink") === "yes";
        if (isDashboardLink && !isQuickLink) {
            anchor.on("click", function (ev) {
                ev.preventDefault();
                progressiveDashboardPageGet($(this).attr("href"), true);
            });
            $(anchor).attr("gyQuickLink", "yes");
        }
    });
    $("form").each(function () {
        var form = $(this);
        var method = form.attr("method");
        var action = $(this).attr("action");
        if (typeof action === "undefined") {
            action = document.location;
        }
        // TODO: support for method GET
        if (method.toUpperCase() !== "POST") {
            return;
        }
        // Upload form doesn't support this yet
        var exclude = $(this).attr("gyExcludeQuickForm");
        if (typeof exclude !== "undefined") {
            return;
        }
        var isDashboardLink = action.includes(getBaseUrl() + "/codeinfest/dashboard");
        var isQuickForm = $(form).attr("gyQuickForm") === "yes";
        if (isDashboardLink && !isQuickForm) {
            form.on("submit", function (ev) {
                ev.preventDefault();
                var action = $(this).attr("action");
                if (typeof action === "undefined") {
                    action = document.location;
                }
                var formData = $(this).serialize();
                progressiveDashboardPagePost(action, formData);
            });
            $(form).attr("gyQuickForm", "yes");
        }
    });
}

function playPing() {
    const playPromise = pingSound.play();
    if (playPromise !== null) {
        playPromise.catch(() => {
            console.log("Cannot play ping sound because Chrome restriction!");
        });
    }
}

function buildNotificationHtml(userName, avatar, description, link, since) {
    var html = "<a href=\"" + getBaseUrl() + "/codeinfest" + link + "\" class=\"list-group-item list-group-item-action\">";
    html += "<div class=\"notification-info\">";
    html += "<div class=\"notification-list-user-img\"><img src=\"" + getBaseUrl() + "/codeinfest" + avatar + "\" alt=\"" + userName + "Avatar\" class=\"user-avatar-md rounded-circle\"></div>";
    html += "<div class=\"notification-list-user-block\"><span class=\"notification-list-user-name\">" + userName + " </span>" + description;
    html += "<div class=\"notification-date\">" + since + "</div>";
    html += "</div></div></a>";
    return html;
}

function refreshAjaxNotifications() {
    $.ajax({
        type: "GET",
        url: getBaseUrl() + "/codeinfest/ajax/getNotifications",
        success: function (resp, status, xhr) {
            var lastUpdate = resp.updateTime;
            if (lastUpdate > notificationLastUpdate) {
                var nthtml = "<p class=\"text-muted text-center\">No New notifications</p>";
                if (Array.isArray(resp.notifications)) {
                    var ntLen = resp.notifications.length;
                    nthtml = "";
                    for (var i = 0; i < ntLen; i++) {
                        var ntObj = resp.notifications[i];
                        nthtml += buildNotificationHtml(ntObj.fromUserName, ntObj.fromUserAvatar, ntObj.description, ntObj.urlLink, ntObj.receivedTimeStr);
                    }
                    if ((resp.updated) && (ntLen > 0)) {
                        playPing();
                    }
                }
                $("#gyNotificationList").html(nthtml);
                notificationLastUpdate = lastUpdate;
            }
            resp.updated && Array.isArray(resp.notifications) ? $("#gyNotificationNew").show() : $("#gyNotificationNew").hide();
            redefineAnchorVisit();
            setTimeout(refreshAjaxNotifications, 5000);
        },
        error: function (xhr, reason, ex) {
            console.log(reason + ": " + ex);
            setTimeout(refreshAjaxNotifications, 5000);
        }
    });
}

$(document).ready(function () {
    // onpopstate
    window.onpopstate = function (event) {
        progressiveDashboardPageGet(document.location, false);
    };
    // Apply OverlayScrollbars and workarounds for BS Modal
    var osInstance = $('body').overlayScrollbars({}).overlayScrollbars();
    $('body').on('show.bs.modal', function () {
        requestAnimationFrame(function () {
            var osContentElm = $(osInstance.getElements().content);
            var backdropElms = $('body > .modal-backdrop');
            backdropElms.each(function (index, elm) {
                osContentElm.append(elm);
            });
        });
    });
    // Notification icon click handle
    $("#gyNotificationIcon").click(function () {
        $.ajax({
            type: "GET",
            url: "/codeinfest/ajax/readAllNotifications",
            success: function (resp, status, xhr) {
                if (resp.succeeded) {
                    $("#gyNotificationNew").hide();
                }
            }
        });
    });
    // About application handle
    $("#aboutAppLink").click(function () {
        var modalHeight = $(window).height() / 2;
        $("#aboutModalContainer").css("height", modalHeight + "px");
        //$('#aboutModalContainer').overlayScrollbars({});
        $.ajax({
            type: "GET",
            url: getBaseUrl() + "/codeinfest/about",
            success: function (resp, status, xhr) {
                $("#aboutModalContainer").html(resp);
            },
            error: function (xhr, reason, ex) {
                alert(reason + ": " + ex);
            }
        });
    });
    dashboardClock();
    $("#gyClockContainer").show();
    $('#onlineOverflow').overlayScrollbars({});
    //
    $(".nav-left-sidebar").hover(function () {
        sidebarHovered = true;
        maximizeSidebar(true);
    }, function () {
        sidebarHovered = false;
        minimizeSidebar(true);
    });
    setTimeout(function () {
        if (!sidebarHovered) {
            minimizeSidebar(true);
        }
    }, 3000);
    // Page-defined custom initialization
    subviewInit();
    redefineAnchorVisit();
    refreshAjaxNotifications();
    animateDashboardMessage();
});