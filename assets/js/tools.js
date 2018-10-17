
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