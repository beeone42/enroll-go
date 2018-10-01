function ldap_promo_from_dn(dn) {
  var re = /ou=([0-9]{4}),/;
  var res = dn.match(re);
  console.log(res);
  if (res != null) {
    return (res[1]);
  }
  return("---");
}

function ldap_uid_from_dn(dn) {
  var re = /uid=([a-z0-9-_]+),/;
  var res = dn.match(re);
  console.log(res);
  if (res != null) {
    return (res[1]);
  }
  return("---");
}

function ldap_by_rfid(q_rfid, success) {
  $("#enroll_rfid").prop('disabled', true);
    $.getJSON("/api/ldap/byrfid/" + q_rfid, function (infos) {
      if (infos == null) {
        warn('LDAP Search RFID returned nothing');
        return ;
      }
      if (infos.length > 0) {
        $("#ldap_cn").text(infos[0]['cn']);
        $("#ldap_dn").text(infos[0]['dn']);
        $("#ldap_rfid").text(infos[0]['badgerfid']);
        $("#ldap_pin").text(infos[0]['badgepin']);
        $("#ldap_promo").text(ldap_promo_from_dn(infos[0]['dn']));
        if (success != null) {
        $("#enroll_rfid").prop('disabled', false);;
          success();
        }
      }
    });
}

function ldap_by_login(q_login, success) {
  $("#enroll_rfid").prop('disabled', true);
    $.getJSON("/api/ldap/bylogin/" + q_login, function (infos) {
      if (infos == null) {
        warn('LDAP Search LOGIN returned nothing');

  tac_by_email(q_login + "@", null);

        return ;
      }
      if (infos.length > 0) {
        $("#ldap_cn").text(infos[0]['cn']);
        $("#ldap_dn").text(infos[0]['dn']);
        $("#ldap_rfid").text(infos[0]['badgerfid']);
        $("#ldap_pin").text(infos[0]['badgepin']);
        $("#ldap_promo").text(ldap_promo_from_dn(infos[0]['dn']));
    if ($("#p_name").text() == "") {
      $("#p_name").text(infos[0]['cn']);
    }
    if ($("#p_company").text() == "") {
      $("#p_company").text(infos[0]['dn'].split(",")[1].split("=")[1]);
    }
      login = ldap_uid_from_dn(infos[0]['dn']);
    $("#avatar").attr("src", photo_url + login);
    if ($("#p_email").text() == "") {
      $("#p_email").text(infos[0]['alias']);
    }
    if ($("#p_rfid").text() == "") {
      $("#p_rfid").text(infos[0]['badgerfid']);
    }
    if (success != null) {
      $("#enroll_rfid").prop('disabled', false);;
      success();
    }
      }
    });
}

function bank_set_refund(login, refund)
{
  $.get("/api/bank/refund/" + login + "/" + refund, function (data) {
    console.log(data);
  });
}

function fill_bank_data(data)
{
  $("#bank_refund").change(function () {});

  $("#bank_balance").text(data['balance']);
  $("#bank_refund" ).prop( "checked", (data["refund"] == "1"));

  $("#bank_refund").change(function () {
    if ($(this).prop("checked"))
    {
      bank_set_refund(data['login'], '1');
      info("refund on");
    } else {
      bank_set_refund(data['login'], '0');
      warn("refund off");
    }
  });
}

function bank_by_rfid(rfid)
{
  $.getJSON("/api/bank/byrfid/" + rfid, function (data) {
    if (data["success"] == "TRUE")
      fill_bank_data(data);
    else
      warn("failed to get bank infos");
  });
}


function fill_tac_data(data) {

  $.getJSON("/api/tac/user/byid/" + data['id'], function (infos) {

    tac_login = infos['email'].split("@")[0];
    if (login != "") {
      if ((tac_login != "") && (tac_login != login))
      {
        warn("TAC Login mismatch: " + login + " != " + tac_login)
        return
      }
    } else {
      login = tac_login;
    }

    $("#p_name").text(data['name']);
    $("#p_company").text(data['company']);

    $("#ca_id").text(data['id']);
    $("#ca_name").text(data['name']);
    $("#ca_company").text(data['company']);

    $("#p_email").text(infos['email']);
    $("#ca_email").text(infos['email']);
    $("#avatar").attr("src", photo_url + login);

    $.getJSON("/api/tac/profile/byid/" + data['id'], function (profiles) {
      //alert(profiles);
      var names = profiles.map(p => p.name);
      $("#ca_profiles").text(names.join(", "));
    });
    $.getJSON("/api/tac/tags/byid/" + data['id'], function (tags) {
      var ids = tags.map(p => formatTag(p.id));
      $("#ca_tags").html(ids.join(", "));
    });

  });

}

function  tac_by_email(q_email, success) {
  if (q_email.length < 4)
  {
    warn('TAC Search LOGIN is too short');
    return ;
  }
    
  $.getJSON("/api/tac/user/byemail/" + q_email, function (data) {
      if ((data == null) || (data[0]['id'] == null)) {
        warn('TAC Search LOGIN returned nothing');
        return ;
      }
      else
      {
  i = 0;
  while (data[i] != null) {
    var j = i;
          $.getJSON("/api/tac/user/byid/" + data[i]['id'], function (infos) {
            if (infos['email'].substring(0, q_email.length) == q_email) {
              fill_tac_data(data[j]);
            }
    });
          i++;
  }
  if (success != null) {
    success();
  }
      }
  });
}

function  tac_by_rfid(q_rfid, success) {
    $.getJSON("/api/tac/user/byrfid/" + q_rfid, function (data) {
      if ((data == null) || (data['id'] == null)) {
        warn('TAC Search RFID returned nothing');
        return ;
      }
      else
      {
        fill_tac_data(data);
    if (success != null) {
        success();
    }
      }
  });
}

function l(l) {
  $("#auto").html("");
  $("#login").val(l);
  submitForm();
}

function auto_complete_sort(a, b) {
  f = 'cn';
  return (a[f] == b[f] ? 0 : (a[f] > b[f] ? 1 : -1));
}

function auto_complete(q) {

  if (q.length < 5) {
    $("#auto").html("need more chars to complete <b>" + q + "</b>");
    return ;
  }

  $.getJSON("/api/ldap/autocomplete/" + q, function (data) {
    if (data != null) {
      data.sort(auto_complete_sort);
      $("#auto").html("");
      for (i = 0; data[i] != null; i++) {
        tmp = data[i]['dn'].split(",")[0].split("=")[1];
        $("#auto").append("<span onclick=\"l('" + tmp + "')\">" + data[i]['cn'] + " <b>" + data[i]['dn'] + "</b></span><br />\n");
      }
    }
    else
    {
  $("#auto").html("no match for <b>" + q + "</b>");
    }
  }).fail(function() {
    $("#auto").html("query failed for <b>" + q + "</b>");
  });
}

function submitForm() {
  if ($("#login").val() != "")
    document.location.href = "/profile/login/" + $("#login").val().trim();
  else
    document.location.href = "/profile/rfid/" + $("#rfid").val().trim();
}

