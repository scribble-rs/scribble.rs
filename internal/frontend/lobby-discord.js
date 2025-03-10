import { DiscordSDK } from '{{.RootPath}}/resources/discord-sdk.mjs';
const clientId = "1320396325925163070"
const discordSdk = new DiscordSDK(clientId, null, {
    frameId: getCookie("discord-frame-id"),
    instanceId: getCookie("discord-instance-id"),
    platform: getCookie("discord-platform"),
});
// postMessage(JSON.parse(window.localStorage.getItem("sdk-payload")));

async function setupDiscordSdk() {
    await discordSdk.ready()

    console.log("Discord SDK ready")
    return discordSdk
        .authorize({
            client_id: clientId,
            response_type: "code",
            state: "",
            prompt: "none",
            scope: [
                "identify",
            ],
        })
        .then((response) => {
            console.log("Authenticating client");
            return fetch("`{{.RootPath}}/v1/discord_authenticate`", {
                method: "POST",
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    code: response.code,
                }),
            })
        })
        .then((response) => response.json())
        .then((response) => {
            return discordSdk
                .authenticate({
                    access_token: response.access_code(),
                })
        })
}

setupDiscordSdk()
    .then((response) => {
        let name = response.user.username;
        if (response.user.nickname) {
            name = response.user.nickname;
        }
        console.log(name);
        changeName(name);
    }).
    catch((error) => {
        console.log(error);
    });

