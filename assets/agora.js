class Video{
    constructor(options = {}){
        let {appID,token,channel,codec,mode} = options;
        ['appID', 'token', 'channel','mode','codec'].forEach(item => {
            this[item] = eval(item)
        });
        this.resolutions = [
            {
                name: 'default',
                value: 'default',
            },
            {
                name: '480p',
                value: '480p',
            },
            {
                name: '720p',
                value: '720p',
            },
            {
                name: '1080p',
                value: '1080p'
            }
        ];
        this.Toast = {
            info: (msg) => {
                this.Toastify({
                    text: msg,
                    classes: "info-toast"
                })
            },
            notice: (msg) => {
                this.Toastify({
                    text: msg,
                    classes: "notice-toast"
                })
            },
            error: (msg) => {
                this.Toastify({
                    text: msg,
                    classes: "error-toast"
                })
            }
        };
        this.rtc = {
            client: null,
            joined: false,
            published: false,
            localStream: null,
            remoteStreams: [],
            params: {}
        };
        this.fields=['appID', 'channel'];
        this.init(this.fields);
    }
      init(){
              this.getDevices((devices)=>{
                  devices.audios.forEach(function (audio) {
                      $('<option/>', {
                          value: audio.value,
                          text: audio.name,
                      }).appendTo("#microphoneId");
                  })
                  devices.videos.forEach(function (video) {
                      $('<option/>', {
                          value: video.value,
                          text: video.name,
                      }).appendTo("#cameraId");
                  })
                  this.resolutions.forEach(function (resolution) {
                      $('<option/>', {
                          value: resolution.value,
                          text: resolution.name
                      }).appendTo("#cameraResolution");
                  })
                  //M.AutoInit();
              })
     /*         var fields = this.fields;
              $("#publish").on("click", (e)=>{
                  console.log("publish");
                  e.preventDefault();
                  var params = this.serializeformData();
                  if (this.validator(params, fields)) {
                      this.publish(this.rtc);
                  }
              });
              $("#unpublish").on("click",(e)=>{
                  console.log("unpublish")
                  e.preventDefault();
                  var params = this.serializeformData();
                  if (this.validator(params, fields)) {
                      this.unpublish(this.rtc);
                  }
              });
              $("#leave").on("click",(e)=>{
                  console.log("leave")
                  e.preventDefault();
                  var params = this.serializeformData();
                  if (this.validator(params, fields)) {
                      this.leave(this.rtc);
                  }
              })
*/
      }

    Toastify (options) {

        /*M.toast({html: options.text, classes: options.classes});*/
    }


     validator(formData, fields) {
        var keys = Object.keys(formData);
        for (let key of keys) {
            if (fields.indexOf(key) != -1) {
                if (!formData[key]) {
                    this.Toast.error("Please Enter " + key);
                    return false;
                }
            }
        }
        return true;
    }

     serializeformData() {
        let formData = $("#form").serializeArray();
        let obj = {};
        for (let item of formData) {
            let key = item.name;
            let val = item.value;
            obj[key] = val;
        }
         obj.appID=this.appID;
         obj.token=this.token;
         obj.channel=this.channel;
         obj.mode=this.mode;
         obj.codec=this.codec;
        return obj;
    }

     addView (id, show) {
        if (!$("#" + id)[0]) {
            $("<div/>", {
                id: "remote_video_panel_" + id,
                class: "video-view",
            }).appendTo("#video");

            $("<div/>", {
                id: "remote_video_" + id,
                class: "video-placeholder",
            }).appendTo("#remote_video_panel_" + id);

            $("<div/>", {
                id: "remote_video_info_" + id,
                class: "video-profile " + (show ? "" :  "hide"),
            }).appendTo("#remote_video_panel_" + id);

            $("<div/>", {
                id: "video_autoplay_"+ id,
                class: "autoplay-fallback hide",
            }).appendTo("#remote_video_panel_" + id);
        }
    }
     removeView (id) {
        if ($("#remote_video_panel_" + id)[0]) {
            $("#remote_video_panel_"+id).remove();
        }
    }
     getDevices (next) {
        AgoraRTC.getDevices(function (items) {
            items.filter(function (item) {
                return ['audioinput', 'videoinput'].indexOf(item.kind) !== -1
            })
                .map(function (item) {
                    return {
                        name: item.label,
                        value: item.deviceId,
                        kind: item.kind,
                    }
                });
            var videos = [];
            var audios = [];
            for (var i = 0; i < items.length; i++) {
                var item = items[i];
                if ('videoinput' == item.kind) {
                    var name = item.label;
                    var value = item.deviceId;
                    if (!name) {
                        name = "camera-" + videos.length;
                    }
                    videos.push({
                        name: name,
                        value: value,
                        kind: item.kind
                    });
                }
                if ('audioinput' == item.kind) {
                    var name = item.label;
                    var value = item.deviceId;
                    if (!name) {
                        name = "microphone-" + audios.length;
                    }
                    audios.push({
                        name: name,
                        value: value,
                        kind: item.kind
                    });
                }
            }
            next({videos: videos, audios: audios});
        });
    }
     handleEvents (rtc) {
        // Occurs when an error message is reported and requires error handling.
        rtc.client.on("error", (err) => {
            console.log(err)
        })
        // Occurs when the peer user leaves the channel; for example, the peer user calls Client.leave.
        rtc.client.on("peer-leave",(evt)=>{
            var id = evt.uid;
            console.log("id", evt);
            if (id != rtc.params.uid) {
                this.removeView(id);
            }
            this.Toast.notice("peer leave")
            console.log('peer-leave', id);
        })
        // Occurs when the local stream is published.
        rtc.client.on("stream-published",(evt)=>{
            this.Toast.notice("stream published success")
            console.log("stream-published");
        })
        // Occurs when the remote stream is added.
        rtc.client.on("stream-added", function (evt) {
            var remoteStream = evt.stream;
            var id = remoteStream.getId();

            if (id !== rtc.params.uid) {
                rtc.client.subscribe(remoteStream, function (err) {
                    console.log("stream subscribe failed", err);
                })
            }
            console.log('stream-added remote-uid: ', id);
        });
        // Occurs when a user subscribes to a remote stream.
        rtc.client.on("stream-subscribed",(evt)=>{
            var remoteStream = evt.stream;
            var id = remoteStream.getId();
            rtc.remoteStreams.push(remoteStream);
            this.addView(id);
            remoteStream.play("remote_video_" + id);

            console.log('stream-subscribed remote-uid: ', id);
        })
        // Occurs when the remote stream is removed; for example, a peer user calls Client.unpublish.
        rtc.client.on("stream-removed",(evt)=>{
            var remoteStream = evt.stream;
            var id = remoteStream.getId();

            remoteStream.stop("remote_video_" + id);
            rtc.remoteStreams = rtc.remoteStreams.filter(function (stream) {
                return stream.getId() !== id
            })
            this.removeView(id);
            console.log('stream-removed remote-uid: ', id);
        })
        rtc.client.on("onTokenPrivilegeWillExpire", function(){
            // After requesting a new token
            // rtc.client.renewToken(token);

            console.log("onTokenPrivilegeWillExpire")
        });
        rtc.client.on("onTokenPrivilegeDidExpire", function(){
            // After requesting a new token
            // client.renewToken(token);

            console.log("onTokenPrivilegeDidExpire")
        })
    }

     join (rtc, option) {

        if (rtc.joined) {
            this.Toast.error("Your already joined");
            return;
        }

        /**
         * A class defining the properties of the config parameter in the createClient method.
         * Note:
         *    Ensure that you do not leave mode and codec as empty.
         *    Ensure that you set these properties before calling Client.join.
         *  You could find more detail here. https://docs.agora.io/en/Video/API%20Reference/web/interfaces/agorartc.clientconfig.html
         **/
        rtc.client = AgoraRTC.createClient({mode: option.mode, codec: option.codec});

        rtc.params = option;

        // handle AgoraRTC client event
        this.handleEvents(rtc);

        // init client
        rtc.client.init(option.appID,  ()=> {
            console.log("init success");

            /**
             * Joins an AgoraRTC Channel
             * This method joins an AgoraRTC channel.
             * Parameters
             * tokenOrKey: string | null
             *    Low security requirements: Pass null as the parameter value.
             *    High security requirements: Pass the string of the Token or Channel Key as the parameter value. See Use Security Keys for details.
             *  channel: string
             *    A string that provides a unique channel name for the Agora session. The length must be within 64 bytes. Supported character scopes:
             *    26 lowercase English letters a-z
             *    26 uppercase English letters A-Z
             *    10 numbers 0-9
             *    Space
             *    "!", "#", "$", "%", "&", "(", ")", "+", "-", ":", ";", "<", "=", ".", ">", "?", "@", "[", "]", "^", "_", "{", "}", "|", "~", ","
             *  uid: number | null
             *    The user ID, an integer. Ensure this ID is unique. If you set the uid to null, the server assigns one and returns it in the onSuccess callback.
             *   Note:
             *      All users in the same channel should have the same type (number or string) of uid.
             *      If you use a number as the user ID, it should be a 32-bit unsigned integer with a value ranging from 0 to (232-1).
             **/
            rtc.client.join(option.token ? option.token : null, option.channel, option.uid ? +option.uid : null,(uid)=>{
                this.Toast.notice("join channel: " + option.channel + " success, uid: " + uid);
                console.log("join channel: " + option.channel + " success, uid: " + uid);
                rtc.joined = true;

                rtc.params.uid = uid;

                // create local stream
                rtc.localStream = AgoraRTC.createStream({
                    streamID: rtc.params.uid,
                    audio: true,
                    video: true,
                    screen: false,
                    microphoneId: option.microphoneId,
                    cameraId: option.cameraId
                })

                // init local stream
                rtc.localStream.init(()=> {
                    console.log("init local stream success");
                    // play stream with html element id "local_stream"
                    rtc.localStream.play("local_stream")

                    // publish local stream
                    this.publish(rtc);
                },(err)=>{
                    this.Toast.error("stream init failed, please open console see more detail")
                    console.error("init local stream failed ", err);
                })
            }, (err)=>{
                this.Toast.error("client join failed, please open console see more detail")
                console.error("client join failed", err)
            })
        }, (err) => {
            this.Toast.error("client init failed, please open console see more detail")
            console.error(err);
        });
    }

     publish (rtc) {
        if (!rtc.client) {
            this.Toast.error("Please Join Room First");
            return;
        }
        if (rtc.published) {
            this.Toast.error("Your already published");
            return;
        }
        var oldState = rtc.published;

        // publish localStream
        rtc.client.publish(rtc.localStream,(err)=>{
            rtc.published = oldState;
            console.log("publish failed");
            this.Toast.error("publish failed")
            console.error(err);
        })
        this.Toast.info("publish")
        rtc.published = true
         console.log(rtc.published);
     }

     unpublish (rtc) {
    if (!rtc.client) {

            return;
        }
        if (!rtc.published) {
            console.log(rtc.published);
            return;
        }

         let oldState = rtc.published;
         console.log(oldState);
         rtc.client.unpublish(rtc.localStream,(err)=>{
            rtc.published = oldState;
            console.log("unpublish failed");
            console.error(err);
        })
        this.Toast.info("unpublish")
        rtc.published = false;
    }

     leave (rtc) {
        if (!rtc.client) {
            this.Toast.error("Please Join First!");
            return;
        }
        if (!rtc.joined) {
            this.Toast.error("You are not in channel");
            return;
        }
        /**
         * Leaves an AgoraRTC Channel
         * This method enables a user to leave a channel.
         **/
        rtc.client.leave(()=>{
            // stop stream
            rtc.localStream.stop();
            // close stream
            rtc.localStream.close();
            while (rtc.remoteStreams.length > 0) {
                var stream = rtc.remoteStreams.shift();
                var id = stream.getId();
                stream.stop();
                this.removeView(id);
            }
            rtc.localStream = null;
            rtc.remoteStreams = [];
            rtc.client = null;
            console.log("client leaves channel success");
            rtc.published = false;
            rtc.joined = false;
            this.Toast.notice("leave success");
        },(err)=>{
            console.log("channel leave failed");
            this.Toast.error("leave success");
            console.error(err);
        })
    }
    joinVideo(){
        console.log(this);
      /*  e.preventDefault();*/
            var params = this.serializeformData();
            if (this.validator(params, this.fields)) {
                this.join(this.rtc, params);
            }
        }
    unpublishVideo(){
            var params = this.serializeformData();
            if (this.validator(params, this.fields)) {
                this.unpublish(this.rtc);
            }
    };
    publishVideo(){
        var params = this.serializeformData();
        if (this.validator(params, this.fields)) {
            this.publish(this.rtc);
        }
    };
    leaveVideo(){
       /*     e.preventDefault();*/
            var params = this.serializeformData();
            if (this.validator(params, this.fields)) {
                this.leave(this.rtc);
            }
    }
}
