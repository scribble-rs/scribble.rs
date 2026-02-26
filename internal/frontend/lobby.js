String.prototype.format = function () {
    return [...arguments].reduce((p, c) => p.replace(/%s/, c), this);
};

const discordInstanceId = getCookie("discord-instance-id");
const rootPath = `${discordInstanceId ? ".proxy/" : ""}{{.RootPath}}`;

let socketIsConnecting = false;
let hasSocketEverConnected = false;
let socket;

const reconnectDialogId = "reconnect-dialog";
function showReconnectDialogIfNotShown() {
    const previousReconnectDialog = document.getElementById(reconnectDialogId);

    //Since the content is constant, there's no need to ever show two.
    if (
        previousReconnectDialog === undefined ||
        previousReconnectDialog === null
    ) {
        showTextDialog(
            reconnectDialogId,
            '{{.Translation.Get "connection-lost"}}',
            `{{.Translation.Get "connection-lost-text"}}`,
        );
    }
}

//Makes sure that the server notices that the player disconnects.
//Otherwise a refresh (on chromium based browsers) can lead to the server
//thinking that there's already an open tab with this lobby.
window.onbeforeunload = () => {
    //Avoid unintentionally reestablishing connection.
    socket.onclose = null;
    if (socket) {
        socket.close();
    }
};

const messageInput = document.getElementById("message-input");
const playerContainer = document.getElementById("player-container");
const wordContainer = document.getElementById("word-container");
const chat = document.getElementById("chat");
const messageContainer = document.getElementById("message-container");
const roundSpan = document.getElementById("rounds");
const maxRoundSpan = document.getElementById("max-rounds");
const timeLeftValue = document.getElementById("time-left-value");
const drawingBoard = document.getElementById("drawing-board");

const centerDialogs = document.getElementById("center-dialogs");

const waitChooseDialog = document.getElementById("waitchoose-dialog");
const waitChooseDrawerSpan = document.getElementById("waitchoose-drawer");
const namechangeDialog = document.getElementById("namechange-dialog");
const namechangeFieldStartDialog = document.getElementById(
    "namechange-field-start-dialog",
);
const namechangeField = document.getElementById("namechange-field");

const lobbySettingsButton = document.getElementById("lobby-settings-button");
const lobbySettingsDialog = document.getElementById("lobbysettings-dialog");

const startDialog = document.getElementById("start-dialog");
const forceStartButton = document.getElementById("force-start-button");
const gameOverDialog = document.getElementById("game-over-dialog");
const gameOverDialogTitle = document.getElementById("game-over-dialog-title");
const gameOverScoreboard = document.getElementById("game-over-scoreboard");
const forceRestartButton = document.getElementById("force-restart-button");
const wordDialog = document.getElementById("word-dialog");
const wordPreSelected = document.getElementById("word-preselected");
const wordButtonContainer = document.getElementById("word-button-container");

const kickDialog = document.getElementById("kick-dialog");
const kickDialogPlayers = document.getElementById("kick-dialog-players");

const soundToggleLabel = document.getElementById("sound-toggle-label");
let sound = localStorage.getItem("sound") !== "false";
updateSoundIcon();

const penToggleLabel = document.getElementById("pen-pressure-toggle-label");
let penPressure = localStorage.getItem("penPressure") !== "false";
updateTogglePenIcon();

function showTextDialog(id, title, message) {
    const messageNode = document.createElement("span");
    messageNode.innerText = message;
    showDialog(id, title, messageNode);
}

const menu = document.getElementById("menu");
function hideMenu() {
    menu.hidePopover();
}

const menu_button = document.getElementById("menu-button");
menu.addEventListener("toggle", (event) => {
    if (event.newState === "open") {
        const bounds = menu_button.getBoundingClientRect();
        menu.style.top = bounds.bottom + "px";

        // making sure the menu doesn't go off-screen
        const menuWidth = menu.offsetWidth;
        const viewportWidth = window.innerWidth;
        if (bounds.left + menuWidth > viewportWidth) {
            menu.style.left = (viewportWidth - menuWidth - 5) + "px";
        } else {
            menu.style.left = bounds.left + "px";
        }
    }
});

function showDialog(id, title, contentNode, buttonBar) {
    hideMenu();

    const newDialog = document.createElement("div");
    newDialog.classList.add("center-dialog");
    if (id && id !== "") {
        newDialog.id = id;
    }

    const dialogTitle = document.createElement("span");
    dialogTitle.classList.add("dialog-title");
    dialogTitle.innerText = title;
    newDialog.appendChild(dialogTitle);

    const dialogContent = document.createElement("div");
    dialogContent.classList.add("center-dialog-content");
    dialogContent.appendChild(contentNode);
    newDialog.appendChild(dialogContent);

    if (buttonBar !== null && buttonBar !== undefined) {
        newDialog.appendChild(buttonBar);
    }

    newDialog.style.visibility = "visible";
    centerDialogs.appendChild(newDialog);
}

// Shows an information dialog with a button that closes the dialog and
// removes it from the DOM.
function showInfoDialog(title, message, buttonText) {
    const dialogId = "info_dialog";
    closeDialog(dialogId);

    const closeButton = createDialogButton(buttonText);
    closeButton.addEventListener("click", () => {
        closeDialog(dialogId);
    });

    const messageNode = document.createElement("span");
    messageNode.innerText = message;

    showDialog(
        dialogId,
        title,
        messageNode,
        createDialogButtonBar(closeButton),
    );
}

function createDialogButton(text) {
    const button = document.createElement("button");
    button.innerText = text;
    button.classList.add("dialog-button");
    return button;
}

function createDialogButtonBar(...buttons) {
    const buttonBar = document.createElement("div");
    buttonBar.classList.add("button-bar");
    buttons.forEach(buttonBar.appendChild);
    return buttonBar;
}

function closeDialog(id) {
    const dialog = document.getElementById(id);
    if (dialog !== undefined && dialog !== null) {
        const parent = dialog.parentElement;
        if (parent !== undefined && parent !== null) {
            parent.removeChild(dialog);
        }
    }
}

const helpDialogId = "help-dialog";
function showHelpDialog() {
    closeDialog(helpDialogId);
    const controlsLabel = document.createElement("b");
    controlsLabel.innerText = '{{.Translation.Get "controls"}}';

    const controlsTextOne = document.createElement("p");
    controlsTextOne.innerText = '{{.Translation.Get "switch-tools-intro"}}:';

    const controlsTextTwo = document.createElement("p");
    controlsTextTwo.innerHTML =
        '{{.Translation.Get "pencil"}}: <kbd>Q</kbd><br/>' +
        '{{.Translation.Get "fill-bucket"}}: <kbd>W</kbd><br/>' +
        '{{.Translation.Get "eraser"}}: <kbd>E</kbd><br/>';

    const controlsTextThree = document.createElement("p");
    controlsTextThree.innerHTML =
        '{{printf (.Translation.Get "switch-pencil-sizes") "<kbd>1</kbd>" "<kbd>4</kbd>"}}';

    const closeButton = createDialogButton('{{.Translation.Get "close"}}');
    closeButton.addEventListener("click", () => {
        closeDialog(helpDialogId);
    });

    const footer = document.createElement("div");
    footer.className = "help-footer";
    footer.innerHTML = `{{template "footer" . }}`;

    const buttonBar = createDialogButtonBar(closeButton);

    const dialogContent = document.createElement("div");
    dialogContent.appendChild(controlsLabel);
    dialogContent.appendChild(controlsTextOne);
    dialogContent.appendChild(controlsTextTwo);
    dialogContent.appendChild(controlsTextThree);
    dialogContent.appendChild(footer);

    showDialog(
        helpDialogId,
        '{{.Translation.Get "help"}}',
        dialogContent,
        buttonBar,
    );
}
document
    .getElementById("help-button")
    .addEventListener("click", showHelpDialog);

function showKickDialog() {
    hideMenu();

    if (cachedPlayers && cachedPlayers) {
        kickDialogPlayers.innerHTML = "";

        cachedPlayers.forEach((player) => {
            //Don't wanna allow kicking ourselves.
            if (player.id !== ownID && player.connected) {
                const playerKickEntry = document.createElement("button");
                playerKickEntry.classList.add("kick-player-button");
                playerKickEntry.classList.add("dialog-button");
                playerKickEntry.onclick = () => onVotekickPlayer(player.id);
                playerKickEntry.innerText = player.name;
                kickDialogPlayers.appendChild(playerKickEntry);
            }
        });

        kickDialog.style.visibility = "visible";
    }
}
document
    .getElementById("kick-button")
    .addEventListener("click", showKickDialog);

function hideKickDialog() {
    kickDialog.style.visibility = "hidden";
}
document
    .getElementById("kick-close-button")
    .addEventListener("click", hideKickDialog);

function showNameChangeDialog() {
    hideMenu();

    namechangeDialog.style.visibility = "visible";
    namechangeField.focus();
}
document
    .getElementById("name-change-button")
    .addEventListener("click", showNameChangeDialog);

function hideNameChangeDialog() {
    namechangeDialog.style.visibility = "hidden";
}
document
    .getElementById("namechange-close-button")
    .addEventListener("click", hideNameChangeDialog);

function changeName(name) {
    //Avoid unnecessary traffic.
    if (name !== ownName) {
        socket.send(
            JSON.stringify({
                type: "name-change",
                data: name,
            }),
        );
    }
}
document
    .getElementById("namechange-button-start-dialog")
    .addEventListener("click", () => {
        changeName(
            document.getElementById("namechange-field-start-dialog").value,
        );
    });
document.getElementById("namechange-button").addEventListener("click", () => {
    changeName(document.getElementById("namechange-field").value);
    hideNameChangeDialog();
});

function setUsernameLocally(name) {
    ownName = name;
    namechangeFieldStartDialog.value = name;
    namechangeField.value = name;
}

function toggleFullscreen() {
    if (document.fullscreenElement !== null) {
        document.exitFullscreen();
    } else {
        document.body.requestFullscreen();
    }
}
document
    .getElementById("toggle-fullscreen-button")
    .addEventListener("click", toggleFullscreen);

function showLobbySettingsDialog() {
    hideMenu();
    lobbySettingsDialog.style.visibility = "visible";
}
lobbySettingsButton.addEventListener("click", showLobbySettingsDialog);

function hideLobbySettingsDialog() {
    lobbySettingsDialog.style.visibility = "hidden";
}
document
    .getElementById("lobby-settings-close-button")
    .addEventListener("click", hideLobbySettingsDialog);

function saveLobbySettings() {
    fetch(
        `${rootPath}/v1/lobby?` +
            new URLSearchParams({
                drawing_time: document.getElementById(
                    "lobby-settings-drawing-time",
                ).value,
                rounds: document.getElementById("lobby-settings-max-rounds")
                    .value,
                public: document.getElementById("lobby-settings-public")
                    .checked,
                max_players: document.getElementById(
                    "lobby-settings-max-players",
                ).value,
                clients_per_ip_limit: document.getElementById(
                    "lobby-settings-clients-per-ip-limit",
                ).value,
                custom_words_per_turn: document.getElementById(
                    "lobby-settings-custom-words-per-turn",
                ).value,
            }),
        {
            method: "PATCH",
        },
    ).then((result) => {
        if (result.status === 200) {
            hideLobbySettingsDialog();
        } else {
            result.text().then((bodyText) => {
                alert(
                    "Error saving lobby settings: \n\n - " +
                        bodyText.replace(";", "\n - "),
                );
            });
        }
    });
}
document
    .getElementById("lobby-settings-save-button")
    .addEventListener("click", saveLobbySettings);

function toggleSound() {
    sound = !sound;
    localStorage.setItem("sound", sound.toString());
    updateSoundIcon();
}
document
    .getElementById("toggle-sound-button")
    .addEventListener("click", toggleSound);

function updateSoundIcon() {
    if (sound) {
        soundToggleLabel.src = `{{.RootPath}}/resources/{{.WithCacheBust "sound.svg"}}`;
    } else {
        soundToggleLabel.src = `{{.RootPath}}/resources/{{.WithCacheBust "no-sound.svg"}}`;
    }
}

function togglePenPressure() {
    penPressure = !penPressure;
    localStorage.setItem("penPressure", penPressure.toString());
    updateTogglePenIcon();
}
document
    .getElementById("toggle-pen-pressure-button")
    .addEventListener("click", togglePenPressure);

function updateTogglePenIcon() {
    if (penPressure) {
        penToggleLabel.src = `{{.RootPath}}/resources/{{.WithCacheBust "pen.svg"}}`;
    } else {
        penToggleLabel.src = `{{.RootPath}}/resources/{{.WithCacheBust "no-pen.svg"}}`;
    }
}

//The drawing board has a base size. This base size results in a certain ratio
//that the actual canvas has to be resized accordingly too. This is needed
//since not every client has the same screensize.
const baseWidth = 1600;
const baseHeight = 900;
const boardRatio = baseWidth / baseHeight;

// Moving this here to extract the context after resizing
const context = drawingBoard.getContext("2d", { alpha: false });

// One might one wonder what the fuck is going here. I'll enlighten you!
// The data you put into a canvas, might not necessarily be what comes out
// of it again. Some browser (*cough* firefox *cough*) seem to put little
// off by one / two errors into the data, when reading it back out.
// Apparently this helps against some type of fingerprinting. In order to
// combat this, we do not use the canvas as a source of truth, but
// permanently hold a virtual canvas buffer that we can operate on when
// filling or drawing.
let imageData;

function scaleUpFactor() {
    return baseWidth / drawingBoard.clientWidth;
}

// Will convert the value to the server coordinate space.
// The canvas locally can be bigger or smaller. Depending on the base
// values and the local values, we'll either have a value slightly
// higher or lower than 1.0. Since we draw on a virtual canvas, we have
// to use the server coordinate space, which then gets scaled by the
// canvas API of the browser, as we have a different clientWidth than
// width and clientHeight than height.
function convertToServerCoordinate(value) {
    return Math.round(parseFloat(scaleUpFactor() * value));
}

const pen = 0;
const rubber = 1;
const fillBucket = 2;

let allowDrawing = false;
let spectating = false;
let spectateRequested = false;

//Initially, we require some values to avoid running into nullpointers
//or undefined errors. The specific values don't really matter.
let localTool = pen;
let localLineWidth = 8;

//Those are not scaled for now, as the whole toolbar would then have to incorrectly size up and down.
const sizeButton8 = document.getElementById("size-8-button");
const sizeButton16 = document.getElementById("size-16-button");
const sizeButton24 = document.getElementById("size-24-button");
const sizeButton32 = document.getElementById("size-32-button");
const sizeButtons = document.getElementById("size-buttons");

const toolButtonPen = document.getElementById("tool-type-pencil");
const toolButtonRubber = document.getElementById("tool-type-rubber");
const toolButtonFill = document.getElementById("tool-type-fill");

if (sizeButton8.checked) {
    setLineWidthNoUpdate(8);
} else if (sizeButton16.checked) {
    setLineWidthNoUpdate(16);
} else if (sizeButton24.checked) {
    setLineWidthNoUpdate(24);
} else if (sizeButton32.checked) {
    setLineWidthNoUpdate(32);
}

if (toolButtonPen.checked) {
    chooseToolNoUpdate(pen);
} else if (toolButtonFill.checked) {
    chooseToolNoUpdate(fillBucket);
} else if (toolButtonRubber.checked) {
    chooseToolNoUpdate(rubber);
}

let localColor, localColorIndex;

function setColor(index) {
    setColorNoUpdate(index);

    // If we select a new color, we assume we don't want to use the
    // rubber anymore and automatically switch to the pen.
    if (localTool === rubber) {
        // Clicking the button programmatically won't trigger its
        toolButtonPen.click();

        // updateDrawingStateUI is implicit
        chooseTool(pen);
    } else {
        updateDrawingStateUI();
    }
}

const firstColorButtonRow = document.getElementById("first-color-button-row");
const secondColorButtonRow = document.getElementById("second-color-button-row");
for (let i = 0; i < firstColorButtonRow.children.length; i++) {
    const _setColor = () => setColor(i);
    firstColorButtonRow.children[i].addEventListener("mousedown", _setColor);
    firstColorButtonRow.children[i].addEventListener("click", _setColor);
}
for (let i = 0; i < secondColorButtonRow.children.length; i++) {
    const _setColor = () => setColor(i + 13);
    secondColorButtonRow.children[i].addEventListener("mousedown", _setColor);
    secondColorButtonRow.children[i].addEventListener("click", _setColor);
}

function setColorNoUpdate(index) {
    localColorIndex = index;
    localColor = indexToRgbColor(index);
    sessionStorage.setItem("local_color", JSON.stringify(index));
}

setColorNoUpdate(
    JSON.parse(sessionStorage.getItem("local_color")) ?? 13 /* black*/,
);
updateDrawingStateUI();

function setLineWidth(value) {
    setLineWidthNoUpdate(value);
    updateDrawingStateUI();
}
sizeButton8.addEventListener("change", () => setLineWidth(8));
document
    .getElementById("size-8-button-wrapper")
    .addEventListener("mouseup", sizeButton8.click);
document
    .getElementById("size-8-button-wrapper")
    .addEventListener("mousedown", sizeButton8.click);
sizeButton16.addEventListener("change", () => setLineWidth(16));
document
    .getElementById("size-16-button-wrapper")
    .addEventListener("mouseup", sizeButton16.click);
document
    .getElementById("size-16-button-wrapper")
    .addEventListener("mousedown", sizeButton16.click);
sizeButton24.addEventListener("change", () => setLineWidth(24));
document
    .getElementById("size-24-button-wrapper")
    .addEventListener("mouseup", sizeButton24.click);
document
    .getElementById("size-24-button-wrapper")
    .addEventListener("mousedown", sizeButton24.click);
sizeButton32.addEventListener("change", () => setLineWidth(32));
document
    .getElementById("size-32-button-wrapper")
    .addEventListener("mouseup", sizeButton32.click);
document
    .getElementById("size-32-button-wrapper")
    .addEventListener("mousedown", sizeButton32.click);

function setLineWidthNoUpdate(value) {
    localLineWidth = value;
}

function chooseTool(value) {
    chooseToolNoUpdate(value);
    updateDrawingStateUI();
}
toolButtonFill.addEventListener("change", () => chooseTool(fillBucket));
toolButtonPen.addEventListener("change", () => chooseTool(pen));
toolButtonRubber.addEventListener("change", () => chooseTool(rubber));
document
    .getElementById("tool-type-fill-wrapper")
    .addEventListener("mouseup", toolButtonFill.click);
document
    .getElementById("tool-type-pencil-wrapper")
    .addEventListener("mouseup", toolButtonPen.click);
document
    .getElementById("tool-type-rubber-wrapper")
    .addEventListener("mouseup", toolButtonRubber.click);
document
    .getElementById("tool-type-fill-wrapper")
    .addEventListener("mousedown", toolButtonFill.click);
document
    .getElementById("tool-type-pencil-wrapper")
    .addEventListener("mousedown", toolButtonPen.click);
document
    .getElementById("tool-type-rubber-wrapper")
    .addEventListener("mousedown", toolButtonRubber.click);

function chooseToolNoUpdate(value) {
    if (value === pen || value === rubber || value === fillBucket) {
        localTool = value;
    } else {
        //If this ends up with an invalid value, we use the pencil.
        localTool = pen;
    }
}

function rgbColorObjectToHexString(color) {
    return (
        "#" +
        numberTo16BitHexadecimal(color.r) +
        numberTo16BitHexadecimal(color.g) +
        numberTo16BitHexadecimal(color.b)
    );
}

function numberTo16BitHexadecimal(number) {
    return Number(number).toString(16).padStart(2, "0");
}

const rubberColor = { r: 255, g: 255, b: 255 };

function updateDrawingStateUI() {
    // Color all buttons, so the player always has a hint as to what the
    // active color is, since the cursor might not always be visible.
    sizeButtons.style.setProperty(
        "--dot-color",
        rgbColorObjectToHexString(localColor),
    );

    updateCursor();
}

function updateCursor() {
    if (allowDrawing) {
        if (localTool === rubber) {
            setCircleCursor(rubberColor, localLineWidth);
        } else if (localTool === fillBucket) {
            const outerColor = getComplementaryCursorColor(localColor);
            drawingBoard.style.cursor =
                `url('data:image/svg+xml;utf8,` +
                encodeURIComponent(
                    `<svg xmlns="http://www.w3.org/2000/svg" version="1.1" height="32" width="32">` +
                        generateSVGCircle(8, localColor, outerColor) +
                        //This has been taken from fill.svg
                        `
                                <svg viewBox="0 0 64 64" x="8" y="8" height="24" width="24">
                                    <path
                                        d="m 59.575359,58.158246 c 0,1.701889 -1.542545,3.094345 -3.427877,3.094345 H 8.1572059 c -1.8853322,0 -3.4278772,-1.392456 -3.4278772,-3.094345 V 5.5543863 c 0,-1.7018892 1.542545,-3.0943445 3.4278772,-3.0943445 H 56.147482 c 1.885332,0 3.427877,1.3924553 3.427877,3.0943445 z"
                                        id="path8"
                                        style="stroke-width:1.62842;fill:#b3b3b3" />
                                    <path
                                        d="M 56.147482,2.4600418 H 8.1572059 c -1.8853322,0 -3.4278772,1.3152251 -3.4278772,2.9227219 V 14.15093 c 0,1.607497 0,0 0,0 l 26.5660453,2.922722 c 0.685576,0 2.570908,0.584545 2.570908,1.89977 0,0 0,1.899769 0,2.484313 0,1.169089 1.199758,2.192042 2.570908,2.192042 1.371151,0 2.570908,-1.022953 2.570908,-2.192042 0,-1.169089 1.199756,-2.192041 2.570908,-2.192041 1.37115,0 2.570907,1.022952 2.570907,2.192041 v 19.728374 c 0,1.169089 1.199757,2.192042 2.570908,2.192042 1.37115,0 2.570907,-1.022953 2.570907,-2.192042 V 25.841818 c 0,-1.169088 1.199758,-2.192041 2.570908,-2.192041 1.371151,0 2.570908,1.022953 2.570908,2.192041 v 3.653404 c 0,1.169088 1.199756,2.192041 2.570907,2.192041 1.371151,0 2.570908,-1.022953 2.570908,-2.192041 V 5.3827637 c 0,-1.6074968 -1.542545,-2.9227219 -3.427877,-2.9227219 z"
                                        id="path12"
                                        style="stroke-width:1.58262;fill:#C75C5C" />
                                    <path
                                        d="m 60.432329,6.1134441 c 0,13.2983859 -12.683145,24.1124579 -28.279986,24.1124579 -15.596839,0 -28.2799836,-10.814072 -28.2799836,-24.1124579"
                                        id="path18"
                                        style="fill:none;stroke:#4F5D73;stroke-width:2;stroke-linecap:round;stroke-miterlimit:10" />
                                </svg>
                            </svg>`,
                ) +
                `') 4 4, auto`;
        } else {
            setCircleCursor(localColor, localLineWidth);
        }
    } else {
        drawingBoard.style.cursor = "not-allowed";
    }
}

function getComplementaryCursorColor(innerColor) {
    const hsp = Math.sqrt(
        0.299 * (innerColor.r * innerColor.r) +
            0.587 * (innerColor.g * innerColor.g) +
            0.114 * (innerColor.b * innerColor.b),
    );

    if (hsp > 127.5) {
        return { r: 0, g: 0, b: 0 };
    }

    return { r: 255, g: 255, b: 255 };
}

function setCircleCursor(innerColor, size) {
    const outerColor = getComplementaryCursorColor(innerColor);
    const circleSize = size;
    drawingBoard.style.cursor =
        `url('data:image/svg+xml;utf8,` +
        encodeURIComponent(
            `<svg xmlns="http://www.w3.org/2000/svg" version="1.1" width="32" height="32">` +
                generateSVGCircle(circleSize, innerColor, outerColor) +
                `</svg>')`,
        ) +
        ` ` +
        circleSize / 2 +
        ` ` +
        circleSize / 2 +
        `, auto`;
}

function generateSVGCircle(circleSize, innerColor, outerColor) {
    const circleRadius = circleSize / 2;
    const innerColorCSS =
        "rgb(" + innerColor.r + "," + innerColor.g + "," + innerColor.b + ")";
    const outerColorCSS =
        "rgb(" + outerColor.r + "," + outerColor.g + "," + outerColor.b + ")";
    return (
        `<circle cx="` +
        circleRadius +
        `" cy="` +
        circleRadius +
        `" r="` +
        circleRadius +
        `" style="fill: ` +
        innerColorCSS +
        `; stroke: ` +
        outerColorCSS +
        `;"/>`
    );
}

function toggleSpectate() {
    socket.send(
        JSON.stringify({
            type: "toggle-spectate",
        }),
    );
}
document
    .getElementById("toggle-spectate-button")
    .addEventListener("click", toggleSpectate);

function setSpectateMode(requestedValue, spectatingValue) {
    const modeUnchanged = spectatingValue === spectating;
    const requestUnchanged = requestedValue === spectateRequested;
    if (modeUnchanged && requestUnchanged) {
        return;
    }

    if (spectateRequested && !requestedValue && modeUnchanged) {
        showInfoDialog(
            `{{.Translation.Get "spectation-request-cancelled-title"}}`,
            `{{.Translation.Get "spectation-request-cancelled-text"}}`,
            `{{.Translation.Get "confirm"}}`,
        );
    } else if (spectateRequested && !requestedValue && modeUnchanged) {
        showInfoDialog(
            `{{.Translation.Get "participation-request-cancelled-title"}}`,
            `{{.Translation.Get "participation-request-cancelled-text"}}`,
            `{{.Translation.Get "confirm"}}`,
        );
    } else if (!spectateRequested && requestedValue && !spectatingValue) {
        showInfoDialog(
            `{{.Translation.Get "spectation-requested-title"}}`,
            `{{.Translation.Get "spectation-requested-text"}}`,
            `{{.Translation.Get "confirm"}}`,
        );
    } else if (!spectateRequested && requestedValue && spectatingValue) {
        showInfoDialog(
            `{{.Translation.Get "participation-requested-title"}}`,
            `{{.Translation.Get "participation-requested-text"}}`,
            `{{.Translation.Get "confirm"}}`,
        );
    } else if (spectatingValue && !spectating) {
        showInfoDialog(
            `{{.Translation.Get "now-spectating-title"}}`,
            `{{.Translation.Get "now-spectating-text"}}`,
            `{{.Translation.Get "confirm"}}`,
        );
    } else if (!spectatingValue && spectating) {
        showInfoDialog(
            `{{.Translation.Get "now-participating-title"}}`,
            `{{.Translation.Get "now-participating-text"}}`,
            `{{.Translation.Get "confirm"}}`,
        );
    }

    spectateRequested = requestedValue;
    spectating = spectatingValue;
}

function toggleReadiness() {
    socket.send(
        JSON.stringify({
            type: "toggle-readiness",
        }),
    );
}
document
    .getElementById("ready-state-start")
    .addEventListener("change", toggleReadiness);
document
    .getElementById("ready-state-game-over")
    .addEventListener("change", toggleReadiness);

function forceStartGame() {
    socket.send(
        JSON.stringify({
            type: "start",
        }),
    );
}
forceStartButton.addEventListener("click", forceStartGame);
forceRestartButton.addEventListener("click", forceStartGame);

function clearCanvasAndSendEvent() {
    if (allowDrawing) {
        //Avoid unnecessary traffic back to us and handle the clear directly.
        clear(context);
        socket.send(
            JSON.stringify({
                type: "clear-drawing-board",
            }),
        );
    }
}
document
    .getElementById("clear-canvas-button")
    .addEventListener("click", clearCanvasAndSendEvent);

function undoAndSendEvent() {
    if (allowDrawing) {
        socket.send(
            JSON.stringify({
                type: "undo",
            }),
        );
    }
}
document
    .getElementById("undo-button")
    .addEventListener("click", undoAndSendEvent);

//Used to restore the last message on arrow up.
let lastMessage = "";

const encoder = new TextEncoder();
function sendMessage(event) {
    if (event.key !== "Enter") {
        return;
    }
    if (!messageInput.value) {
        return;
    }

    // While the backend already checks for message length, we want to
    // prevent the loss of input and omit the event / clear here.
    if (encoder.encode(messageInput.value).length > 10000) {
        appendMessage(
            "system-message",
            '{{.Translation.Get "system"}}',
            '{{.Translation.Get "message-too-long"}}',
        );
        //We keep the messageInput content, since it could've been
        //something important and we don't want the user having to
        //rewrite it. Instead they can send it via some other means
        //or shorten it a bit.
        return;
    }

    socket.send(
        JSON.stringify({
            type: "message",
            data: messageInput.value,
        }),
    );
    lastMessage = messageInput.value;
    messageInput.value = "";
}

messageInput.addEventListener("keypress", sendMessage);
messageInput.addEventListener("keydown", function (event) {
    if (event.key === "ArrowUp" && messageInput.value.length === 0) {
        messageInput.value = lastMessage;
        const length = lastMessage.length;
        // Postpone selection change onto next event queue loop iteration, as
        // nothing will happen otherwise.
        setTimeout(() => {
            // length+1 is necessary, as the selection wont change if start and
            // end are the same,
            messageInput.setSelectionRange(length + 1, length);
        }, 0);
    }
});

function setAllowDrawing(value) {
    allowDrawing = value;
    updateDrawingStateUI();

    if (allowDrawing) {
        document.getElementById("toolbox").style.display = "flex";
    } else {
        document.getElementById("toolbox").style.display = "none";
    }
}

function chooseWord(index) {
    socket.send(
        JSON.stringify({
            type: "choose-word",
            data: index,
        }),
    );
    setAllowDrawing(true);
    wordDialog.style.visibility = "hidden";
}

function onVotekickPlayer(playerId) {
    socket.send(
        JSON.stringify({
            type: "kick-vote",
            data: playerId,
        }),
    );
    hideKickDialog();
}

//This automatically scrolls down the chat on arrivals of new messages
new MutationObserver(
    () => (messageContainer.scrollTop = messageContainer.scrollHeight),
).observe(messageContainer, {
    attributes: false,
    childList: true,
    subtree: false,
});

let ownID, ownerID, ownName, drawerID, drawerName;
let round = 0;
let rounds = 0;
let roundEndTime = 0;
let gameState = "unstarted";
let drawingTimeSetting = "∞";

const handleEvent = (parsed) => {
    if (parsed.type === "ready") {
        handleReadyEvent(parsed.data);
    } else if (parsed.type === "game-over") {
        let ready = parsed.data;
        if (parsed.data.roundEndReason === "drawer_disconnected") {
            appendMessage(
                "system-message",
                null,
                `{{.Translation.Get "drawer-disconnected"}}`,
            );
        } else if (parsed.data.roundEndReason === "guessers_disconnected") {
            appendMessage(
                "system-message",
                null,
                `{{.Translation.Get "guessers-disconnected"}}`,
            );
        } else {
            showRoundEndMessage(ready.previousWord);
        }
        handleReadyEvent(ready);
    } else if (parsed.type === "update-players") {
        applyPlayers(parsed.data);
    } else if (parsed.type === "name-change") {
        const player = getCachedPlayer(parsed.data.playerId);
        if (player !== null) {
            player.name = parsed.data.playerName;
        }

        const playernameSpan = document.getElementById(
            "playername-" + parsed.data.playerId,
        );
        if (playernameSpan !== null) {
            playernameSpan.innerText = parsed.data.playerName;
        }
        if (parsed.data.playerId === ownID) {
            setUsernameLocally(parsed.data.playerName);
        }
        if (parsed.data.playerId === drawerID) {
            waitChooseDrawerSpan.innerText = parsed.data.playerName;
        }
    } else if (parsed.type === "correct-guess") {
        playWav('{{.RootPath}}/resources/{{.WithCacheBust "plop.wav"}}');

        if (parsed.data === ownID) {
            appendMessage(
                "correct-guess-message",
                null,
                `{{.Translation.Get "correct-guess"}}`,
            );
        } else {
            const player = getCachedPlayer(parsed.data);
            if (player !== null) {
                appendMessage(
                    "correct-guess-message-other-player",
                    null,
                    `{{.Translation.Get "correct-guess-other-player"}}`.format(
                        player.name,
                    ),
                );
            }
        }
    } else if (parsed.type === "close-guess") {
        appendMessage(
            "close-guess-message",
            null,
            `{{.Translation.Get "close-guess"}}`.format(parsed.data),
        );
    } else if (parsed.type === "update-wordhint") {
        wordDialog.style.visibility = "hidden";
        waitChooseDialog.style.visibility = "hidden";
        applyWordHints(parsed.data);

        // We don't do this in applyWordHints because that's called in all kinds of places
        if (parsed.data.some((hint) => hint.character)) {
            appendMessage(
                "system-message",
                '{{.Translation.Get "system"}}',
                `{{.Translation.Get "word-hint-revealed"}}`.format(
                    parsed.data
                        .map((hint) =>
                            hint.character && hint.revealed
                                ? String.fromCharCode(hint.character)
                                : "_",
                        )
                        .join(" "),
                ),
            );
        }
    } else if (parsed.type === "message") {
        appendMessage(null, parsed.data.author, parsed.data.content);
    } else if (parsed.type === "system-message") {
        appendMessage(
            "system-message",
            '{{.Translation.Get "system"}}',
            parsed.data,
        );
    } else if (parsed.type === "non-guessing-player-message") {
        appendMessage(
            "non-guessing-player-message",
            parsed.data.author,
            parsed.data.content,
        );
    } else if (parsed.type === "line") {
        drawLine(
            context,
            imageData,
            parsed.data.x,
            parsed.data.y,
            parsed.data.x2,
            parsed.data.y2,
            indexToRgbColor(parsed.data.color),
            parsed.data.width,
        );
    } else if (parsed.type === "fill") {
        if (
            floodfillUint8ClampedArray(
                imageData.data,
                parsed.data.x,
                parsed.data.y,
                indexToRgbColor(parsed.data.color),
                imageData.width,
                imageData.height,
            )
        ) {
            context.putImageData(imageData, 0, 0);
        }
    } else if (parsed.type === "clear-drawing-board") {
        clear(context);
    } else if (parsed.type === "word-chosen") {
        wordDialog.style.visibility = "hidden";
        waitChooseDialog.style.visibility = "hidden";
        setRoundTimeLeft(parsed.data.timeLeft);
        applyWordHints(parsed.data.hints);
        setAllowDrawing(drawerID === ownID);
    } else if (parsed.type === "next-turn") {
        if (gameState === "ongoing") {
            if (parsed.data.roundEndReason === "drawer_disconnected") {
                appendMessage(
                    "system-message",
                    null,
                    `{{.Translation.Get "drawer-disconnected"}}`,
                );
            } else if (parsed.data.roundEndReason === "guessers_disconnected") {
                appendMessage(
                    "system-message",
                    null,
                    `{{.Translation.Get "guessers-disconnected"}}`,
                );
            } else {
                showRoundEndMessage(parsed.data.previousWord);
            }
        } else {
            //First turn, the game starts
            gameState = "ongoing";
        }

        //As soon as a turn starts, the round should be ongoing, so we make
        //sure that all types of dialogs, that indicate the game isn't
        //ongoing, are not visible anymore.
        startDialog.style.visibility = "hidden";
        forceRestartButton.style.display = "none";
        gameOverDialog.style.visibility = "hidden";

        //If a player doesn't choose, the dialog will still be up.
        wordDialog.style.visibility = "hidden";
        playWav('{{.RootPath}}/resources/{{.WithCacheBust "end-turn.wav"}}');

        clear(context);

        round = parsed.data.round;
        updateRoundsDisplay();
        setRoundTimeLeft(parsed.data.choiceTimeLeft);
        applyPlayers(parsed.data.players);

        set_dummy_word_hints();

        //Even though we always hide the dialog in the "your-turn"
        //event handling, it will be shortly visible if we it here.
        if (drawerID !== ownID) {
            //Show additional dialog, that another user (drawer) is choosing a word
            waitChooseDrawerSpan.innerText = drawerName;
            waitChooseDialog.style.visibility = "visible";
        }

        setAllowDrawing(false);
    } else if (parsed.type === "your-turn") {
        playWav('{{.RootPath}}/resources/{{.WithCacheBust "your-turn.wav"}}');
        //This dialog could potentially stay visible from last
        //turn, in case nobody has chosen a word.
        waitChooseDialog.style.visibility = "hidden";
        promptWords(parsed.data);
    } else if (parsed.type === "drawing") {
        applyDrawData(parsed.data);
    } else if (parsed.type === "kick-vote") {
        if (
            parsed.data.playerId === ownID &&
            parsed.data.voteCount >= parsed.data.requiredVoteCount
        ) {
            alert('{{.Translation.Get "self-kicked"}}');
            document.location.href = "{{.RootPath}}/";
        } else {
            let kickMessage = '{{.Translation.Get "kick-vote"}}'.format(
                parsed.data.voteCount,
                parsed.data.requiredVoteCount,
                parsed.data.playerName,
            );
            if (parsed.data.voteCount >= parsed.data.requiredVoteCount) {
                kickMessage += ' {{.Translation.Get "player-kicked"}}';
            }
            appendMessage(
                "system-message",
                '{{.Translation.Get "system"}}',
                kickMessage,
            );
        }
    } else if (parsed.type === "owner-change") {
        ownerID = parsed.data.playerId;
        updateButtonVisibilities();
        appendMessage(
            "system-message",
            '{{.Translation.Get "system"}}',
            '{{.Translation.Get "owner-change"}}'.format(
                parsed.data.playerName,
            ),
        );
    } else if (parsed.type === "drawer-kicked") {
        appendMessage(
            "system-message",
            '{{.Translation.Get "system"}}',
            '{{.Translation.Get "drawer-kicked"}}',
        );
    } else if (parsed.type === "lobby-settings-changed") {
        rounds = parsed.data.rounds;
        updateRoundsDisplay();
        updateButtonVisibilities();
        appendMessage(
            "system-message",
            '{{.Translation.Get "system"}}',
            '{{.Translation.Get "lobby-settings-changed"}}\n\n' +
                '{{.Translation.Get "drawing-time-setting"}}: ' +
                parsed.data.drawingTime +
                "\n" +
                '{{.Translation.Get "rounds-setting"}}: ' +
                parsed.data.rounds +
                "\n" +
                '{{.Translation.Get "public-lobby-setting"}}: ' +
                parsed.data.public +
                "\n" +
                '{{.Translation.Get "max-players-setting"}}: ' +
                parsed.data.maxPlayers +
                "\n" +
                '{{.Translation.Get "custom-words-per-turn-setting"}}: ' +
                parsed.data.customWordsPerTurn +
                "%\n" +
                '{{.Translation.Get "players-per-ip-limit-setting"}}: ' +
                parsed.data.clientsPerIpLimit,
        );
    } else if (parsed.type === "shutdown") {
        socket.onclose = null;
        socket.close();
        showDialog(
            "shutdown-info",
            '{{.Translation.Get "server-shutting-down-title"}}',
            document.createTextNode(
                '{{.Translation.Get "server-shutting-down-text"}}',
            ),
        );
    }
};

function showRoundEndMessage(previousWord) {
    if (previousWord === "") {
        appendMessage(
            "system-message",
            null,
            `{{.Translation.Get "round-over"}}`,
        );
    } else {
        appendMessage(
            "system-message",
            null,
            `{{.Translation.Get "round-over-no-word"}}`.format(previousWord),
        );
    }
}

function getCachedPlayer(playerID) {
    if (!cachedPlayers) {
        return null;
    }

    for (let i = 0; i < cachedPlayers.length; i++) {
        const player = cachedPlayers[i];
        if (player.id === playerID) {
            return player;
        }
    }

    return null;
}

//In the initial implementation we used a timestamp to know when
//the round will end. The problem with that approach was that the
//clock on client and server was often not in sync. The second
//approach was to instead send milliseconds left and keep counting
//them down each 500 milliseconds. The problem with this approach, was
//that there could potentially be timing mistakes while counting down.
//What we do instead is use our local date, add the timeLeft to it and
//repeatdly recaculate the timeLeft using the roundEndTime and the
//current time. This way we won't have any calculation errors.
//
//FIXME The only leftover issue is that ping isn't taken into
//account, however, that's no biggie for now.
function setRoundTimeLeft(timeLeftMs) {
    roundEndTime = Date.now() + timeLeftMs;
}

const handleReadyEvent = (ready) => {
    ownerID = ready.ownerId;
    ownID = ready.playerId;

    setRoundTimeLeft(ready.timeLeft);
    setUsernameLocally(ready.playerName);
    setAllowDrawing(ready.allowDrawing);
    round = ready.round;
    rounds = ready.rounds;
    gameState = ready.gameState;
    drawingTimeSetting = ready.drawingTimeSetting;
    updateRoundsDisplay();
    updateButtonVisibilities();

    if (ready.players && ready.players.length) {
        applyPlayers(ready.players);
    }
    if (ready.currentDrawing && ready.currentDrawing.length) {
        applyDrawData(ready.currentDrawing);
    }
    if (ready.wordHints && ready.wordHints.length) {
        applyWordHints(ready.wordHints);
    } else {
        set_dummy_word_hints();
    }

    if (ready.gameState === "unstarted") {
        startDialog.style.visibility = "visible";
        if (ownerID === ownID) {
            forceStartButton.style.display = "block";
        } else {
            forceStartButton.style.display = "none";
        }
    } else if (ready.gameState === "gameOver") {
        gameOverDialog.style.visibility = "visible";
        if (ownerID === ownID) {
            forceRestartButton.style.display = "block";
        }

        gameOverScoreboard.innerHTML = "";

        //Copying array so we can sort.
        const players = cachedPlayers.slice();
        players.sort((a, b) => {
            return a.rank - b.rank;
        });

        //These two are required for displaying the "game over / win / tie" message.
        let countOfRankOnePlayers = 0;
        let selfPlayer;
        for (let i = 0; i < players.length; i++) {
            const player = players[i];
            if (!player.connected || player.state === "spectating") {
                continue;
            }

            if (player.rank === 1) {
                countOfRankOnePlayers++;
            }
            if (player.id === ownID) {
                selfPlayer = player;
            }

            // We only display the first 5 players on the scoreboard.
            if (player.rank <= 5) {
                const newScoreboardEntry = document.createElement("div");
                newScoreboardEntry.classList.add("gameover-scoreboard-entry");
                if (player.id === ownID) {
                    newScoreboardEntry.classList.add(
                        "gameover-scoreboard-entry-self",
                    );
                }

                const scoreboardRankDiv = document.createElement("div");
                scoreboardRankDiv.classList.add("gameover-scoreboard-rank");
                scoreboardRankDiv.innerText = player.rank;
                newScoreboardEntry.appendChild(scoreboardRankDiv);

                const scoreboardNameDiv = document.createElement("div");
                scoreboardNameDiv.classList.add("gameover-scoreboard-name");
                scoreboardNameDiv.innerText = player.name;
                newScoreboardEntry.appendChild(scoreboardNameDiv);

                const scoreboardScoreSpan = document.createElement("span");
                scoreboardScoreSpan.classList.add("gameover-scoreboard-score");
                scoreboardScoreSpan.innerText = player.score;
                newScoreboardEntry.appendChild(scoreboardScoreSpan);

                gameOverScoreboard.appendChild(newScoreboardEntry);
            }
        }

        if (selfPlayer.rank === 1) {
            if (countOfRankOnePlayers >= 2) {
                gameOverDialogTitle.innerText = `{{.Translation.Get "game-over-tie"}}`;
            } else {
                gameOverDialogTitle.innerText = `{{.Translation.Get "game-over-win"}}`;
            }
        } else {
            gameOverDialogTitle.innerText =
                `{{.Translation.Get "game-over"}}`.format(
                    selfPlayer.rank,
                    selfPlayer.score,
                );
        }
    } else if (ready.gameState === "ongoing") {
        // Lack of wordHints implies that word has been chosen yet.
        if (!ready.wordHints && drawerID !== ownID) {
            waitChooseDrawerSpan.innerText = drawerName;
            waitChooseDialog.style.visibility = "visible";
        }
    }
};

function updateButtonVisibilities() {
    if (ownerID === ownID) {
        lobbySettingsButton.style.display = "flex";
    } else {
        lobbySettingsButton.style.display = "none";
    }
}

function promptWords(data) {
    wordPreSelected.textContent = data.words[data.preSelectedWord];
    wordButtonContainer.replaceChildren(
        ...data.words.map((word, index) => {
            const button = createDialogButton(word);
            button.onclick = () => {
                chooseWord(index);
            };
            return button;
        }),
    );
    wordDialog.style.visibility = "visible";
}

function playWav(file) {
    if (sound) {
        const audio = new Audio(file);
        audio.type = "audio/wav";
        audio.play();
    }
}

window.setInterval(() => {
    if (gameState === "ongoing") {
        const msLeft = roundEndTime - Date.now();
        const secondsLeft = Math.max(0, Math.floor(msLeft / 1000));
        timeLeftValue.innerText = "" + secondsLeft;
    } else {
        timeLeftValue.innerText = "∞";
    }
}, 500);

//appendMessage adds a new message to the message container. If the
//message amount is too high, we cut off a part of the messages to
//prevent lagging and useless memory usage.
function appendMessage(styleClass, author, message) {
    if (messageContainer.childElementCount >= 100) {
        messageContainer.removeChild(messageContainer.firstChild);
    }

    const newMessageDiv = document.createElement("div");
    newMessageDiv.classList.add("message");
    if (styleClass !== null && styleClass !== "") {
        newMessageDiv.classList.add(styleClass);
    }

    if (author !== null && author !== "") {
        const authorNameSpan = document.createElement("span");
        authorNameSpan.classList.add("chat-name");
        authorNameSpan.innerText = author;
        newMessageDiv.appendChild(authorNameSpan);
    }

    const messageSpan = document.createElement("span");
    messageSpan.classList.add("message-content");
    messageSpan.innerText = message;
    newMessageDiv.appendChild(messageSpan);

    messageContainer.appendChild(newMessageDiv);
}

let cachedPlayers;

//applyPlayers takes the players passed, assigns them to cachedPlayers,
//refreshes the scoreboard and updates the drawerID and drawerName variables.
function applyPlayers(players) {
    const matchOngoing = gameState === "ongoing";
    if (!matchOngoing) {
        let readyPlayers = 0;
        let readyPlayersRequired = 0;

        players.forEach((player) => {
            if (!player.connected || player.state === "spectating") {
                return;
            }

            readyPlayersRequired = readyPlayersRequired + 1;
            if (player.state === "ready") {
                readyPlayers = readyPlayers + 1;
            }

            if (player.id === ownID) {
                document.getElementById("ready-state-start").checked =
                    player.state === "ready";
                document.getElementById("ready-state-game-over").checked =
                    player.state === "ready";
            }
        });

        const readyCounts = document.getElementsByClassName("ready-count");
        const reaadyNeededs = document.getElementsByClassName("ready-needed");

        Array.from(readyCounts).forEach((element) => {
            element.innerText = readyPlayers.toString();
        });
        Array.from(reaadyNeededs).forEach((element) => {
            element.innerText = readyPlayersRequired.toString();
        });
    }

    playerContainer.innerHTML = "";
    players.forEach((player) => {
        // Makes sure that the "is choosing" a word dialog doesn't show
        // "undefined" as the player name. Can happen, if the player
        // disconnects after being assigned the drawer.
        if (matchOngoing && player.state === "drawing") {
            drawerID = player.id;
            drawerName = player.name;
        }

        //We don't wanna show the disconnected players.
        if (!player.connected) {
            return;
        }

        if (player.id === ownID) {
            setSpectateMode(
                player.spectateToggleRequested,
                player.state === "spectating",
            );
        }

        const oldPlayer = getCachedPlayer(player.id);
        if (
            oldPlayer &&
            oldPlayer.state === "spectating" &&
            player.state !== "spectating"
        ) {
            appendMessage(
                "system-message",
                '{{.Translation.Get "system"}}',
                `${player.name} is now participating`,
            );
        } else if (
            oldPlayer &&
            oldPlayer.state !== "spectating" &&
            player.state === "spectating"
        ) {
            appendMessage(
                "system-message",
                '{{.Translation.Get "system"}}',
                `${player.name} is now spectating`,
            );
        }

        if (player.state === "spectating") {
            return;
        }

        const playerDiv = document.createElement("div");

        playerDiv.classList.add("player");

        const scoreAndStatusDiv = document.createElement("div");
        scoreAndStatusDiv.classList.add("score-and-status");
        playerDiv.appendChild(scoreAndStatusDiv);

        const playerscoreDiv = document.createElement("div");
        playerscoreDiv.classList.add("playerscore-group");
        scoreAndStatusDiv.appendChild(playerscoreDiv);

        if (matchOngoing) {
            if (player.state === "standby") {
                playerDiv.classList.add("player-done");
            } else if (player.state === "drawing") {
                const playerStateImage = createPlayerStateImageNode(
                    `{{.RootPath}}/resources/{{.WithCacheBust "pencil.svg"}}`,
                );
                playerStateImage.style.transform = "scaleX(-1)";
                scoreAndStatusDiv.appendChild(playerStateImage);
            } else if (player.state === "standby") {
                const playerStateImage = createPlayerStateImageNode(
                    `{{.RootPath}}/resources/{{.WithCacheBust "checkmark.svg"}}`,
                );
                scoreAndStatusDiv.appendChild(playerStateImage);
            }
        } else {
            if (player.state === "ready") {
                playerDiv.classList.add("player-ready");
            }
        }

        const rankSpan = document.createElement("span");
        rankSpan.classList.add("rank");
        rankSpan.innerText = player.rank;
        playerDiv.appendChild(rankSpan);

        const playernameSpan = document.createElement("span");
        playernameSpan.classList.add("playername");
        playernameSpan.innerText = player.name;
        playernameSpan.id = "playername-" + player.id;
        if (player.id === ownID) {
            playernameSpan.classList.add("playername-self");
        }
        playerDiv.appendChild(playernameSpan);

        const playerscoreSpan = document.createElement("span");
        playerscoreSpan.classList.add("playerscore");
        playerscoreSpan.innerText = player.score;
        playerscoreDiv.appendChild(playerscoreSpan);

        const lastPlayerscoreSpan = document.createElement("span");
        lastPlayerscoreSpan.classList.add("last-turn-score");
        lastPlayerscoreSpan.innerText =
            '{{.Translation.Get "last-turn"}}'.format(player.lastScore);
        playerscoreDiv.appendChild(lastPlayerscoreSpan);

        playerContainer.appendChild(playerDiv);
    });

    // We do this at the end, so we can access the old values while
    // iterating over the new ones
    cachedPlayers = players;
}

function createPlayerStateImageNode(path) {
    const playerStateImage = document.createElement("img");
    playerStateImage.style.width = "1rem";
    playerStateImage.style.height = "1rem";
    playerStateImage.src = path;
    return playerStateImage;
}
function updateRoundsDisplay() {
    roundSpan.innerText = round;
    maxRoundSpan.innerText = rounds;
}

const applyWordHints = (wordHints, dummy) => {
    const isDrawer = drawerID === ownID;

    // We abuse the container to prevent the layout from jumping.
    if (!dummy) {
        wordContainer.style.visibility = "visible";
    } else {
        wordContainer.style.visibility = "hidden";
    }

    var wordLengths = [];
    var count = 0;

    wordContainer.replaceChildren(
      ...wordHints.map((hint, index) => {
            const hintSpan = document.createElement("span");
            hintSpan.classList.add("hint");
            if (hint.character === 0) {
                hintSpan.classList.add("hint-underline");
                hintSpan.innerHTML = "&nbsp;";
            } else {
                if (hint.underline) {
                    hintSpan.classList.add("hint-underline");
                }
                hintSpan.innerText = String.fromCharCode(hint.character);
            }

            if (hint.character === 32) {
              wordLengths.push(count);
              count = 0;
            } else if (index === wordHints.length - 1) {
              count += 1;
              wordLengths.push(count);
            } else {
              count += 1;
            }

            if (hint.revealed && isDrawer) {
                hintSpan.classList.add("hint-revealed");
            }

            return hintSpan;
        }),
    );

    const lengthHint = document.createElement("sub");
    lengthHint.classList.add("word-length-hint");
    lengthHint.setAttribute("dir", wordContainer.getAttribute("dir"));
    lengthHint.innerText = `(${wordLengths.join(", ")})`;
    wordContainer.appendChild(lengthHint);
};

const set_dummy_word_hints = () => {
    // Dummy wordhint to prevent layout changes.
    applyWordHints(
        [
            {
                character: "D",
                underline: true,
            },
        ],
        true,
    );
};
set_dummy_word_hints();

const applyDrawData = (drawElements) => {
    clear(context);

    drawElements.forEach((drawElement) => {
        const drawData = drawElement.data;
        if (drawElement.type === "fill") {
            floodfillUint8ClampedArray(
                imageData.data,
                drawData.x,
                drawData.y,
                indexToRgbColor(drawData.color),
                imageData.width,
                imageData.height,
            );
        } else if (drawElement.type === "line") {
            drawLineNoPut(
                context,
                imageData,
                drawData.x,
                drawData.y,
                drawData.x2,
                drawData.y2,
                indexToRgbColor(drawData.color),
                drawData.width,
            );
        } else {
            console.log("Unknown draw element type: " + drawData.type);
        }
    });

    context.putImageData(imageData, 0, 0);
};

let lastX = 0;
let lastY = 0;

let touchID = null;

function onTouchStart(event) {
    //We only allow a single touch
    if (allowDrawing && touchID == null && localTool !== fillBucket) {
        const touch = event.touches[0];
        touchID = touch.identifier;

        // calculate the offset coordinates based on client touch position and drawing board client origin
        const clientRect = drawingBoard.getBoundingClientRect();
        lastX = touch.clientX - clientRect.left;
        lastY = touch.clientY - clientRect.top;
    }
}

function onTouchMove(event) {
    // Prevent moving, scrolling or zooming the page
    event.preventDefault();

    if (allowDrawing) {
        for (let i = event.changedTouches.length - 1; i >= 0; i--) {
            if (event.changedTouches[i].identifier === touchID) {
                const touch = event.changedTouches[i];

                // calculate the offset coordinates based on client touch position and drawing board client origin
                const clientRect = drawingBoard.getBoundingClientRect();
                const offsetX = touch.clientX - clientRect.left;
                const offsetY = touch.clientY - clientRect.top;

                // drawing functions must check for context boundaries
                drawLineAndSendEvent(context, lastX, lastY, offsetX, offsetY);
                lastX = offsetX;
                lastY = offsetY;

                return;
            }
        }
    }
}

function onTouchEnd(event) {
    for (let i = event.changedTouches.length - 1; i >= 0; i--) {
        if (event.changedTouches[i].identifier === touchID) {
            touchID = null;
            return;
        }
    }
}

drawingBoard.addEventListener("touchend", onTouchEnd);
drawingBoard.addEventListener("touchcancel", onTouchEnd);
drawingBoard.addEventListener("touchstart", onTouchStart);
drawingBoard.addEventListener("touchmove", onTouchMove);

function onMouseDown(event) {
    if (
        allowDrawing &&
        event.pointerType !== "touch" &&
        event.buttons === 1 &&
        localTool !== fillBucket
    ) {
        const clientRect = drawingBoard.getBoundingClientRect();
        lastX = event.clientX - clientRect.left;
        lastY = event.clientY - clientRect.top;
    }
}

function pressureToLineWidth(event) {
    //event.button === 0 could be wrong, as it can also be the uninitialized state.
    //Therefore we use event.buttons, which works differently.
    if (
        event.buttons !== 1 ||
        event.pressure === 0 ||
        event.pointerType === "touch"
    ) {
        return 0;
    }
    if (!penPressure || event.pressure === 0.5 || !event.pressure) {
        return localLineWidth;
    }
    return Math.ceil(event.pressure * 32);
}

// Previously the onMouseMove handled leave, but we do this separately now for
// proper pen support. Otherwise leave leads to a loss of the pen pressure, as
// we are handling that with mouseleave instead of pointerleave. pointerlave
// is not triggered until the pen is let go.
function onMouseLeave(event) {
    if (allowDrawing && lastLineWidth && localTool !== fillBucket) {
        // calculate the offset coordinates based on client mouse position and drawing board client origin
        const clientRect = drawingBoard.getBoundingClientRect();
        const offsetX = event.clientX - clientRect.left;
        const offsetY = event.clientY - clientRect.top;

        // drawing functions must check for context boundaries
        drawLineAndSendEvent(
            context,
            lastX,
            lastY,
            offsetX,
            offsetY,
            lastLineWidth,
        );
        lastX = offsetX;
        lastY = offsetY;
    }
}

let lastLineWidth;
function onMouseMove(event) {
    const pressureLineWidth = pressureToLineWidth(event);
    lastLineWidth = pressureLineWidth;

    if (allowDrawing && pressureLineWidth && localTool !== fillBucket) {
        // calculate the offset coordinates based on client mouse position and drawing board client origin
        const clientRect = drawingBoard.getBoundingClientRect();
        const offsetX = event.clientX - clientRect.left;
        const offsetY = event.clientY - clientRect.top;

        // drawing functions must check for context boundaries
        drawLineAndSendEvent(
            context,
            lastX,
            lastY,
            offsetX,
            offsetY,
            pressureLineWidth,
        );
        lastX = offsetX;
        lastY = offsetY;
    }
}

function onMouseClick(event) {
    //event.buttons won't work here, since it's always 0. Since we
    //have a click event, we can be sure that we actually had a button
    //clicked and 0 won't be the uninitialized state.
    if (allowDrawing && event.button === 0) {
        if (localTool === fillBucket) {
            fillAndSendEvent(
                context,
                event.offsetX,
                event.offsetY,
                localColorIndex,
            );
        } else {
            drawLineAndSendEvent(
                context,
                event.offsetX,
                event.offsetY,
                event.offsetX,
                event.offsetY,
            );
        }
    }
}

drawingBoard.addEventListener("pointerdown", onMouseDown);
drawingBoard.addEventListener("pointermove", onMouseMove);
drawingBoard.addEventListener("mouseleave", onMouseLeave);
drawingBoard.addEventListener("click", onMouseClick);

function onGlobalMouseMove(event) {
    const clientRect = drawingBoard.getBoundingClientRect();
    lastX = Math.min(
        clientRect.width - 1,
        Math.max(0, event.clientX - clientRect.left),
    );
    lastY = Math.min(
        clientRect.height - 1,
        Math.max(0, event.clientY - clientRect.top),
    );
}

//necessary for mousemove to not use the previous exit coordinates.
//If this is done via mouseleave and mouseenter of the
//drawingBoard, the lines will end too early on leave and start
//too late on exit.
window.addEventListener("mousemove", onGlobalMouseMove);

function isAnyDialogVisible() {
    for (let i = 0; i < centerDialogs.children.length; i++) {
        if (centerDialogs.children[i].style.visibility === "visible") {
            return true;
        }
    }

    return false;
}

function onKeyDown(event) {
    //Avoid firing actions if the user is in the chat.
    if (document.activeElement instanceof HTMLInputElement) {
        return;
    }

    //If dialogs are open, it doesn't really make sense to be able to
    //change tools. As this is like being in the pause menu of a game.
    if (isAnyDialogVisible()) {
        return;
    }

    //They key choice was made like this, since it's easy to remember
    //and easy to reach. This is how many MOBAs do it and I personally
    //find it better than having to find specific keys on your
    //keyboard. Especially for people that aren't used to typing
    //without looking at their keyboard, this might help.
    if (event.key === "q") {
        toolButtonPen.click();
        chooseTool(pen);
    } else if (event.key === "w") {
        toolButtonFill.click();
        chooseTool(fillBucket);
    } else if (event.key === "e") {
        toolButtonRubber.click();
        chooseTool(rubber);
    } else if (event.key === "1") {
        sizeButton8.click();
        setLineWidth(8);
    } else if (event.key === "2") {
        sizeButton16.click();
        setLineWidth(16);
    } else if (event.key === "3") {
        sizeButton24.click();
        setLineWidth(24);
    } else if (event.key === "4") {
        sizeButton32.click();
        setLineWidth(32);
    } else if (event.key === "z" && event.ctrlKey) {
        undoAndSendEvent();
    }
}

//Handling events on the canvas directly isn't possible, since the user
//must've clicked it at least once in order for that to work.
window.addEventListener("keydown", onKeyDown);

function debounce(func, timeout) {
    let timer;
    return (...args) => {
        clearTimeout(timer);
        timer = setTimeout(() => {
            func.apply(this, args);
        }, timeout);
    };
}

function clear(context) {
    context.fillStyle = "#FFFFFF";
    context.fillRect(0, 0, drawingBoard.width, drawingBoard.height);
    // Refetch, as we don't manually fill here.
    imageData = context.getImageData(
        0,
        0,
        context.canvas.width,
        context.canvas.height,
    );
}

// Clear initially, as it will be black otherwise.
clear(context);

function fillAndSendEvent(context, x, y, colorIndex) {
    const xScaled = convertToServerCoordinate(x);
    const yScaled = convertToServerCoordinate(y);
    const color = indexToRgbColor(colorIndex);
    if (
        floodfillUint8ClampedArray(
            imageData.data,
            xScaled,
            yScaled,
            color,
            imageData.width,
            imageData.height,
        )
    ) {
        context.putImageData(imageData, 0, 0);
        const fillInstruction = {
            type: "fill",
            data: {
                x: xScaled,
                y: yScaled,
                color: colorIndex,
            },
        };
        socket.send(JSON.stringify(fillInstruction));
    }
}

function drawLineAndSendEvent(
    context,
    x1,
    y1,
    x2,
    y2,
    lineWidth = localLineWidth,
) {
    const color = localTool === rubber ? rubberColor : localColor;
    const colorIndex = localTool === rubber ? 0 /* white */ : localColorIndex;

    const x1Scaled = convertToServerCoordinate(x1);
    const y1Scaled = convertToServerCoordinate(y1);
    const x2Scaled = convertToServerCoordinate(x2);
    const y2Scaled = convertToServerCoordinate(y2);
    drawLine(
        context,
        imageData,
        x1Scaled,
        y1Scaled,
        x2Scaled,
        y2Scaled,
        color,
        lineWidth,
    );

    const drawInstruction = {
        type: "line",
        data: {
            x: x1Scaled,
            y: y1Scaled,
            x2: x2Scaled,
            y2: y2Scaled,
            color: colorIndex,
            width: lineWidth,
        },
    };
    socket.send(JSON.stringify(drawInstruction));
}

function getCookie(name) {
    let cookie = {};
    document.cookie.split(";").forEach(function (el) {
        let split = el.split("=");
        cookie[split[0].trim()] = split.slice(1).join("=");
    });
    return cookie[name];
}

const connectToWebsocket = () => {
    if (socketIsConnecting === true) {
        return;
    }

    socketIsConnecting = true;

    socket = new WebSocket(`${rootPath}/v1/lobby/ws`);

    socket.onerror = (error) => {
        //Is not connected and we haven't yet said that we are done trying to
        //connect, this means that we could never even establish a connection.
        if (socket.readyState != 1 && !hasSocketEverConnected) {
            socketIsConnecting = false;
            showTextDialog(
                "connection-error-dialog",
                '{{.Translation.Get "error-connecting"}}',
                `{{.Translation.Get "error-connecting-text"}}`,
            );
            console.log("Error establishing connection: ", error);
        } else {
            console.log("Socket error: ", error);
        }
    };

    socket.onopen = () => {
        closeDialog(reconnectDialogId);

        hasSocketEverConnected = true;
        socketIsConnecting = false;

        socket.onclose = (event) => {
            //We want to avoid handling the error multiple times and showing the incorrect dialogs.
            socket.onerror = null;

            console.log("Socket Closed Connection: ", event);

            if (event.code === 4000) {
                showTextDialog(
                    reconnectDialogId,
                    "Kicked",
                    `You have been kicked from the lobby.`,
                );
            } else {
                console.log("Attempting to reestablish socket connection.");
                showReconnectDialogIfNotShown();
                connectToWebsocket();
            }
        };

        socket.onmessage = (jsonMessage) => {
            handleEvent(JSON.parse(jsonMessage.data));
        };

        console.log("Successfully Connected");
    };
};

connectToWebsocket();

//In order to avoid automatically canceling the socket connection, we keep
//sending dummy events every 5 seconds. This was a problem on Heroku. If
//a player took a very long time to choose a word, the connection of all
//players could be killed and even cause the lobby being closed. Since
//that's very frustrating, we want to avoid that.
window.setInterval(() => {
    if (socket) {
        socket.send(JSON.stringify({ type: "keep-alive" }));
    }
}, 5000);
