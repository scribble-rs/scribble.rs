const discordInstanceId = getCookie("discord-instance-id");
const rootPath = `${discordInstanceId ? ".proxy/" : ""}{{.RootPath}}`;

// Replace native <select> dropdowns with a custom styled widget. The original
// <select> stays in the DOM (hidden) so form submission keeps working.
(function initCustomSelects() {
    const selects = document.querySelectorAll("#lobby-create select");
    selects.forEach(enhanceSelect);

    function enhanceSelect(select) {
        const wrapper = document.createElement("div");
        wrapper.className = "custom-select";

        const button = document.createElement("button");
        button.type = "button";
        button.className = "custom-select-button";
        button.setAttribute("aria-haspopup", "listbox");
        button.setAttribute("aria-expanded", "false");

        const label = document.createElement("span");
        label.className = "custom-select-label";
        button.appendChild(label);

        const chevron = document.createElement("span");
        chevron.className = "custom-select-chevron";
        chevron.setAttribute("aria-hidden", "true");
        button.appendChild(chevron);

        const list = document.createElement("ul");
        list.className = "custom-select-list";
        list.setAttribute("role", "listbox");
        list.hidden = true;

        Array.from(select.options).forEach((opt) => {
            const li = document.createElement("li");
            li.className = "custom-select-option";
            li.setAttribute("role", "option");
            li.dataset.value = opt.value;
            li.textContent = opt.label || opt.text;
            if (opt.title) li.title = opt.title;
            if (opt.selected) li.classList.add("selected");
            li.addEventListener("click", () => {
                selectValue(opt.value);
                closeList();
                button.focus();
            });
            list.appendChild(li);
        });

        function updateLabel() {
            const selected = select.options[select.selectedIndex];
            label.textContent = selected
                ? selected.label || selected.text
                : "";
        }

        function selectValue(value) {
            select.value = value;
            select.dispatchEvent(new Event("change", { bubbles: true }));
            updateLabel();
            list.querySelectorAll(".custom-select-option").forEach((li) => {
                li.classList.toggle("selected", li.dataset.value === value);
            });
        }

        function openList() {
            list.hidden = false;
            button.setAttribute("aria-expanded", "true");
            const selectedLi = list.querySelector(
                ".custom-select-option.selected",
            );
            if (selectedLi)
                selectedLi.scrollIntoView({ block: "nearest" });
            document.addEventListener("mousedown", onOutside, true);
            document.addEventListener("keydown", onKey, true);
        }

        function closeList() {
            list.hidden = true;
            button.setAttribute("aria-expanded", "false");
            document.removeEventListener("mousedown", onOutside, true);
            document.removeEventListener("keydown", onKey, true);
        }

        function onOutside(e) {
            if (!wrapper.contains(e.target)) closeList();
        }

        function onKey(e) {
            if (e.key === "Escape") {
                closeList();
                button.focus();
                return;
            }
            if (e.key === "ArrowDown" || e.key === "ArrowUp") {
                e.preventDefault();
                const items = Array.from(
                    list.querySelectorAll(".custom-select-option"),
                );
                const currentIdx = items.findIndex((li) =>
                    li.classList.contains("selected"),
                );
                const nextIdx =
                    e.key === "ArrowDown"
                        ? Math.min(items.length - 1, currentIdx + 1)
                        : Math.max(0, currentIdx - 1);
                selectValue(items[nextIdx].dataset.value);
                items[nextIdx].scrollIntoView({ block: "nearest" });
            }
        }

        button.addEventListener("click", () => {
            if (list.hidden) openList();
            else closeList();
        });

        updateLabel();

        select.parentNode.insertBefore(wrapper, select);
        wrapper.appendChild(select);
        wrapper.appendChild(button);
        wrapper.appendChild(list);
    }
})();

Array.from(document.getElementsByClassName("number-input")).forEach(
    (number_input) => {
        const input = number_input.children.item(1);
        const decrement_button = number_input.children.item(0);
        decrement_button.addEventListener("click", function () {
            input.stepDown();
        });
        const increment_button = number_input.children.item(2);
        increment_button.addEventListener("click", function () {
            input.stepUp();
        });
    },
);

// We'll keep using the ssr endpoint for now. With this listener, we
// can fake in the form data for "public" depending on which button
// we submitted via. This is a dirty hack, but works for now.
document.getElementById("lobby-create").addEventListener("submit", (event) => {
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

const lobby_list_placeholder = document.getElementById(
    "lobby-list-placeholder-text",
);
const lobby_list_loading_placeholder = document.getElementById(
    "lobby-list-placeholder-loading",
);
const lobby_list = document.getElementById("lobby-list");

lobby_list_placeholder.innerHTML =
    '<b>{{.Translation.Get "no-lobbies-yet"}}</b>';

const getLobbies = () => {
    return new Promise((resolve, reject) => {
        fetch(`${rootPath}/v1/lobby`)
            .then((response) => {
                response.json().then(resolve);
            })
            .catch(reject);
    });
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
        case "hebrew":
            return "\u{1f1ee}\u{1f1f1}";
    }
};

const set_lobbies = (lobbies, visible) => {
    const new_lobby_nodes = lobbies.map((lobby) => {
        const lobby_list_item = document.createElement("div");
        lobby_list_item.className = "lobby-list-item";

        const language_flag = document.createElement("span");
        language_flag.className = "language-flag";
        language_flag.setAttribute("title", lobby.wordpack);
        language_flag.setAttribute("english", lobby.wordpack);
        language_flag.innerText = language_to_flag(lobby.wordpack);

        const lobby_list_rows = document.createElement("div");
        lobby_list_rows.className = "lobby-list-rows";

        const lobby_list_row_a = document.createElement("div");
        lobby_list_row_a.className = "lobby-list-row";

        const new_custom_tag = (text) => {
            const tag = document.createElement("span");
            tag.className = "custom-tag";
            tag.innerText = text;
            return tag;
        };
        if (lobby.customWords) {
            lobby_list_row_a.appendChild(
                new_custom_tag('{{.Translation.Get "custom-words"}}'),
            );
        }
        if (lobby.state === "ongoing") {
            lobby_list_row_a.appendChild(
                new_custom_tag('{{.Translation.Get "ongoing"}}'),
            );
        }
        if (lobby.state === "gameover") {
            lobby_list_row_a.appendChild(
                new_custom_tag('{{.Translation.Get "game-over-lobby"}}'),
            );
        }

        if (lobby.scoring === "chill") {
            lobby_list_row_a.appendChild(
                new_custom_tag('{{.Translation.Get "chill"}}'),
            );
        } else if (lobby.scoring === "competitive") {
            lobby_list_row_a.appendChild(
                new_custom_tag('{{.Translation.Get "competitive"}}'),
            );
        }

        const lobby_list_row_b = document.createElement("div");
        lobby_list_row_b.className = "lobby-list-row";

        const create_info_pair = (icon, text) => {
            const element = document.createElement("div");
            element.className = "lobby-list-item-info-pair";

            const image = document.createElement("img");
            image.className = "lobby-list-item-icon lobby-list-icon-loading";
            image.setAttribute("loading", "lazy");
            image.addEventListener("load", function () {
                image.classList.remove("lobby-list-icon-loading");
            });
            image.setAttribute("src", icon);

            const span = document.createElement("span");
            span.innerText = text;

            element.replaceChildren(image, span);
            return element;
        };
        const user_pair = create_info_pair(
            `{{.RootPath}}/resources/{{.WithCacheBust "user.svg"}}`,
            `${lobby.playerCount}/${lobby.maxPlayers}`,
        );
        const round_pair = create_info_pair(
            `{{.RootPath}}/resources/{{.WithCacheBust "round.svg"}}`,
            `${lobby.round}/${lobby.rounds}`,
        );
        const time_pair = create_info_pair(
            `{{.RootPath}}/resources/{{.WithCacheBust "clock.svg"}}`,
            `${lobby.drawingTime}`,
        );

        lobby_list_row_b.replaceChildren(user_pair, round_pair, time_pair);

        lobby_list_rows.replaceChildren(lobby_list_row_a, lobby_list_row_b);

        const join_button = document.createElement("button");
        join_button.className = "join-button";
        join_button.innerText = '{{.Translation.Get "join"}}';
        join_button.addEventListener("click", (event) => {
            window.location.href = `{{.RootPath}}/lobby/${lobby.lobbyId}`;
        });

        lobby_list_item.replaceChildren(
            language_flag,
            lobby_list_rows,
            join_button,
        );

        return lobby_list_item;
    });
    lobby_list.replaceChildren(...new_lobby_nodes);

    if (lobbies && lobbies.length > 0 && visible) {
        lobby_list.style.display = "flex";
        set_lobby_list_placeholder("", false);
    } else {
        lobby_list.style.display = "none";
        set_lobby_list_placeholder(
            '{{.Translation.Get "no-lobbies-yet"}}',
            true,
        );
    }
};

const refresh_lobby_list = () => {
    set_lobbies([], false);
    set_lobby_list_loading(true);

    getLobbies()
        .then((data) => {
            set_lobbies(data, true);
        })
        .catch((err) => {
            set_lobby_list_placeholder(err, true);
        })
        .finally(() => {
            set_lobby_list_loading(false);
        });
};

refresh_lobby_list();
document
    .getElementById("refresh-lobby-list-button")
    .addEventListener("click", refresh_lobby_list);

function getCookie(name) {
    let cookie = {};
    document.cookie.split(";").forEach(function (el) {
        let split = el.split("=");
        cookie[split[0].trim()] = split.slice(1).join("=");
    });
    return cookie[name];
}

// Makes sure, that navigating back after creating a lobby also shows it in the list.
window.addEventListener("pageshow", (event) => {
    if (event.persisted) {
        refresh_lobby_list();
    }
});
