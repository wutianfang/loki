
var aGET = new Array();
function init() {
    var aQuery = window.location.href.split("?");//取得Get参数
    if(aQuery.length > 1)
    {
        var aBuf = aQuery[1].split("&");
        for(var i=0, iLoop = aBuf.length; i<iLoop; i++)
        {
            var aTmp = aBuf[i].split("=");//分离key与Value
            aGET[aTmp[0]] = aTmp[1];
        }
    }
}

init()


function escapeHTMLString(str) {
    str = str.replace(/</g,'&lt;');
    str = str.replace(/>/g,'&gt;');
    return str;
}

function playAudio(mp3_file) {
    //var url = '/WRecite/audio/' + type + '/' + word.slice(0,2) + '/' + word + '.mp3';
    var audio = document.createElement("audio");
    old && old.pause(),
        audio.src = mp3_file,
        old = audio,
        audio.play();
}
old = null