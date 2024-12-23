const discordInstanceId = getCookie("discord-instance-id")
const rootPath = `${discordInstanceId ? ".proxy/" : ""}{{.RootPath}}`

Array
    .from(document.getElementsByClassName("number-input"))
    .forEach(number_input => {
        const input = number_input.children.item(1);
        const decrement_button = number_input.children.item(0);
        decrement_button.addEventListener("click", function() {
            input.stepDown();
        })
        const increment_button = number_input.children.item(2);
        increment_button.addEventListener("click", function() {
            input.stepUp();
        })
    })

// We'll keep using the ssr endpoint for now. With this listener, we
// can fake in the form data for "public" depending on which button
// we submitted via. This is a dirty hack, but works for now.
document
    .getElementById("lobby-create")
    .addEventListener("submit", (event) => {
        const check_box = document.getElementById("public-check-box");
        if (event.submitter.id === "create-public") {
            check_box.value = "true";
            check_box.setAttribute("checked", "");
        } else {
            check_box.value = "false";
            check_box.removeAttribute("checked");
        }

        return true;
    });

const lobby_list_placeholder =
    document.getElementById("lobby-list-placeholder-text");
const lobby_list_loading_placeholder =
    document.getElementById("lobby-list-placeholder-loading");
const lobby_list = document.getElementById("lobby-list");

lobby_list_placeholder.innerHTML = "<b>There are no lobbies yet.</b>";

const getLobbies = () => {
    return new Promise((resolve, reject) => {
        fetch(`${rootPath}/v1/lobby`).
            then((response) => {
                response.json().then(resolve);
            }).
            catch(reject);
    })
};

const set_lobby_list_placeholder = (text, visible) => {
    if (visible) {
        lobby_list_placeholder.style.display = "flex";
        lobby_list_placeholder.innerHTML = "<b>" + text + "<b>";
    } else {
        lobby_list_placeholder.style.display = "none";
    }
};

const set_lobby_list_loading = (loading) => {
    if (loading) {
        set_lobby_list_placeholder("", false);
        lobby_list_loading_placeholder.style.display = "flex";
    } else {
        lobby_list_loading_placeholder.style.display = "none";
    }
};

const language_to_flag = (language) => {
    switch (language) {
        case "english":
            return "\u{1f1fa}\u{1f1f8}";
        case "english_gb":
            return "\u{1f1ec}\u{1f1e7}";
        case "german":
            return "\u{1f1e9}\u{1f1ea}";
        case "ukrainian":
            return "\u{1f1fa}\u{1f1e6}";
        case "russian":
            return "\u{1f1f7}\u{1f1fa}";
        case "italian":
            return "\u{1f1ee}\u{1f1f9}";
        case "french":
            return "\u{1f1eb}\u{1f1f7}";
        case "dutch":
            return "\u{1f1f3}\u{1f1f1}";
        case "polish":
            return "\u{1f1f5}\u{1f1f1}";
    }
};

const remove_icon_loading_class = (img) => {
    img.classList.remove("lobby-list-icon-loading");
};

const set_lobbies = (lobbies, visible) => {
    const new_lobby_nodes = lobbies.map((lobby) => {
        const lobby_list_item = document.createElement("div");
        lobby_list_item.className = "lobby-list-item";

        const lobby_list_rows = document.createElement("div");
        lobby_list_rows.className = "lobby-list-rows";

        const lobby_list_row_a = document.createElement("div");
        lobby_list_row_a.className = "lobby-list-row";

        const language_flag = document.createElement("span");
        language_flag.className = "language-flag";
        language_flag.innerText = language_to_flag(lobby.wordpack);
        lobby_list_row_a.appendChild(language_flag);

        const new_custom_tag = (text) => {
            const tag = document.createElement("span");
            tag.className = "custom-tag";
            tag.innerText = text;
            return tag;
        };
        if (lobby.customWords) {
            lobby_list_row_a.appendChild(new_custom_tag(
                '{{.Translation.Get "custom-words"}}'
            ));
        }
        if (lobby.state === "ongoing") {
            lobby_list_row_a.appendChild(new_custom_tag(
                'Ongoing'
            ));
        }
        if (lobby.state === "gameover") {
            lobby_list_row_a.appendChild(new_custom_tag(
                'Game Over'
            ));
        }

        if (lobby.scoring === "chill") {
            lobby_list_row_a.appendChild(new_custom_tag(
                '{{.Translation.Get "chill"}}'
            ));
        } else if (lobby.scoring === "competitive") {
            lobby_list_row_a.appendChild(new_custom_tag(
                '{{.Translation.Get "competitive"}}'
            ));
        }

        const lobby_list_row_b = document.createElement("div");
        lobby_list_row_b.className = "lobby-list-row";

        const create_info_pair = (icon, text) => {
            const element = document.createElement("div");
            element.className = "lobby-list-item-info-pair";

            const image = document.createElement("img");
            image.className = "lobby-list-item-icon lobby-list-icon-loading";
            image.setAttribute("loading", "lazy");
            image.setAttribute("onLoad",
                "remove_icon_loading_class(this)");
            image.setAttribute("src", icon);

            const span = document.createElement("span");
            span.innerText = text;

            element.replaceChildren(image, span);
            return element;
        };
        const user_pair = create_info_pair(
            "{{.RootPath}}/resources/user.svg?cache_bust={{.CacheBust}}",
            `${lobby.playerCount}/${lobby.maxPlayers}`);
        const round_pair = create_info_pair(
            "{{.RootPath}}/resources/round.svg?cache_bust={{.CacheBust}}",
            `${lobby.round}/${lobby.rounds}`);
        const time_pair = create_info_pair(
            "{{.RootPath}}/resources/clock.svg?cache_bust={{.CacheBust}}",
            `${lobby.drawingTime}`);

        lobby_list_row_b.replaceChildren(user_pair, round_pair, time_pair);

        lobby_list_rows.replaceChildren(lobby_list_row_a, lobby_list_row_b);

        const join_button = document.createElement("button");
        join_button.className = "join-button";
        join_button.innerText = "Join";
        join_button.addEventListener("click", (event) => {
            window.location.href =
                `{{.RootPath}}/ssrEnterLobby/${lobby.lobbyId}`;
        });

        lobby_list_item.replaceChildren(lobby_list_rows, join_button);

        return lobby_list_item;
    });
    lobby_list.replaceChildren(...new_lobby_nodes);

    if (lobbies && lobbies.length > 0 && visible) {
        lobby_list.style.display = "flex";
        set_lobby_list_placeholder("", false);
    } else {
        lobby_list.style.display = "none";
        set_lobby_list_placeholder("There are no lobbies.", true);
    }
};

const refresh_lobby_list = () => {
    set_lobbies([], false);
    set_lobby_list_loading(true);

    getLobbies().then((data) => {
        set_lobbies(data, true);
    }).catch((err) => {
        set_lobby_list_placeholder(err, true);
    }).finally(() => {
        set_lobby_list_loading(false);
    });
};

refresh_lobby_list();
document.getElementById("refresh-lobby-list-button").addEventListener("click", refresh_lobby_list);

function getCookie(name) {
    let cookie = {};
    document.cookie.split(';').forEach(function(el) {
        let split = el.split('=');
        cookie[split[0].trim()] = split.slice(1).join("=");
    })
    return cookie[name];
}

