const videoValue=  {
    ad:"a1e7f7e62ce3401397d2c7cba836979b",
    tn:"0065f05c8ce1679426aac75f20b14e83bfcIABrPDLSShCjkpxKyxOfPeC4Xvv2YzF5+RIw3gw/7uz/GPfI+dsAAAAAEAAYpZEVBSajXgEAAQDnJaNe",
    cl:"1111",
    me:"rtc",
    cc:"h264"
};
const video=new Video(videoValue);
document.querySelector('#join').addEventListener('click',()=>{
    video.joinVideo()
});
document.querySelector('#leave').addEventListener('click',()=>{
    video.leaveVideo()
});
document.querySelector('#publish').addEventListener('click',
    ()=>{video.publishVideo()
});
document.querySelector('#unpublish').addEventListener('click',()=>{
    video.unpublishVideo()
});