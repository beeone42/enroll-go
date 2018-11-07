
function show_message(type, msg)
{
  $.notify({
    // options
    message: msg
  },{
    // settings
    type: type,
    allow_dismiss: true,
    placement: {
      from: "bottom",
      align: "center"
    }
  });
}

function warn(msg)
{
  show_message("warning", msg)
}

function info(msg)
{
  show_message("info", msg)
}

function success(msg)
{
  show_message("success", msg)
}

function formatTag(tag) {
  if (tag.length == 14) {
    return (tag.substring(0, 10) + "<b>" + tag.substring(10, 14) + "</b>");
  }
  return (tag);
}

function formatDate(ts) {
  var date = new Date(ts * 1000);
  var year = date.getFullYear();
  var month = "0" + (date.getMonth() + 1);
  var day = "0" + date.getDate();
  var formattedDate = year + "-" + month.substr(-2) + "-" + day.substr(-2)

  var hours = date.getHours();
  var minutes = "0" + date.getMinutes();
  var seconds = "0" + date.getSeconds();
  var formattedTime = hours + ':' + minutes.substr(-2) + ':' + seconds.substr(-2);
  return formattedDate + " " + formattedTime
}

function reverseTag(tag) {
  res = ""
  for (i = 0; i < tag.length; i += 2) {
    res = tag.substring(i, i + 2) + res;
  }
  return res
}

function checkSession() {
  $("body").hide();
  token = sessionStorage.getItem('bearer');
  $.ajaxSetup({headers: {'Authorization' : 'Bearer ' + token}});
  $.getJSON("/api/check", function (res) {
    if (res["authentified"] == "false") {
      go(res["goto"]);
    }
  });
  $("body").show();
}

function go(dst) {
  document.location.href = dst;
}