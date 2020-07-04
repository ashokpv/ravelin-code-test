var begin;
var delta = 200;
var end;
var hasTyped = false;
var originalHeight = $(window).height();
var originalWidth = $(window).width();
var acutalHeight = $(window).height();
var actualWidth = $(window).width();
var rtime;
let uuid = uuidv4()
var req = {
  url: window.location.href,
  id: uuid,
  resizeTo: {},
  resizeFrom: {},
};
var timeout = false;

$(document).ready(function() {
  $(window).resize(function() {
    rtime = new Date();
    if (timeout === false) {
      timeout = true;
      setTimeout(resizeend, delta);
    }
  });

  $("input").on('paste', function() {
    req.eventType = "copyAndPaste"
    req.formId = this.id;
    req.pasted = true;
    sendData('copyandpaste');
  })

  $("input").on('copy', function() {
    req.eventType = "copyAndPaste";
    req.formId = this.id;
    req.pasted = false;
    sendData('copyandpaste');
  })

  typeInput();
  submitData();
})

sendData('new');

function resizeend() {
  if (new Date() - rtime < delta) {
    setTimeout(resizeend, delta);
  } else {
    timeout = false;
    screenResolution();
  }
}

function screenResolution() {
  req.eventType = "resize";
  req.resizeFrom.height = originalHeight.toString();
  req.resizeFrom.width = originalWidth.toString();
  actualHeight = $(window).height();
  actualWidth = $(window).width();
  req.resizeTo.height = actualHeight.toString();
  req.resizeTo.width = actualWidth.toString();
  sendData('resize');
  originalHeight = actualHeight;
  originalWidth = actualWidth;
}

function typeInput() {
  $("input").keypress(function(event) {
    if (!hasTyped) {
      hasTyped = true;
      begin = new Date();
    }
  });
}

function submitData() {
  $("button").click(function(event) {
    if (hasTyped) {
      end = new Date();
      interval = (end - begin) / 1000;
      req.eventType = "timeTaken";
      req.interval = interval;
      hasTyped = false;
      sendData('submit');
    }
  });
}

function sendData(route) {
  $.ajax({
    url: 'http://localhost:7080/' + route,
    type: 'POST',
    dataType: 'application/json',
    data: JSON.stringify(
    {
        "websiteURL": req.url,
        "sessionId": req.id.toString(),
        "eventType": (req.eventType) ? req.eventType : null,
        "paste": (req.eventType == 'copyAndPaste') ? req.pasted : null,
        "formId": (req.eventType == 'copyAndPaste') ? req.formId : null,
        "resizeFrom": (req.eventType == 'resize') ? {
          "height": req.resizeFrom.height,
          "width": req.resizeFrom.width
        } : null,
        "resizeTo": (req.eventType == 'resize') ? {
          "height": req.resizeTo.height,
          "width": req.resizeTo.width,
        } : null,
        "time": (req.eventType == 'timeTaken') ? req.interval : null
     }
  )
  })
  .done(function() {
    console.log("success");
  })
  .fail(function() {
    console.log("error");
  })
  .always(function() {
    console.log("complete");
  });
}
