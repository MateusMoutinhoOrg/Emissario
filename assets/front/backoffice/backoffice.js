// The behaviour of the backoffice pages. It lives in this file, served from
// the server itself, because their Content-Security-Policy runs no script
// written inline: neither a <script> block nor an on* attribute.
(function () {
  // The language picker of every page: choosing a language asks the same
  // page again with ?lang=, which the server keeps in a cookie for the next
  // pages; the parameter is then taken out of the address bar.
  document.querySelectorAll('select[data-language]').forEach(function (select) {
    select.addEventListener('change', function () {
      var url = new URL(location.href);
      url.searchParams.set('lang', select.value);
      location.href = url.toString();
    });
  });
  if (new URL(location.href).searchParams.has('lang') && window.history.replaceState) {
    var clean = new URL(location.href);
    clean.searchParams.delete('lang');
    window.history.replaceState(null, '', clean.toString());
  }

  // A form carrying data-confirm asks before it is sent.
  document.querySelectorAll('form[data-confirm]').forEach(function (form) {
    form.addEventListener('submit', function (event) {
      if (!window.confirm(form.dataset.confirm)) {
        event.preventDefault();
      }
    });
  });

  // A submit button carrying data-confirm asks before it sends its form — one
  // form whose buttons do different things, the riskiest one asking.
  document.querySelectorAll('button[data-confirm]').forEach(function (button) {
    button.addEventListener('click', function (event) {
      if (!window.confirm(button.dataset.confirm)) {
        event.preventDefault();
      }
    });
  });

  // The contact requests: a button of the selection does nothing until a row
  // is selected, and the header checkbox selects every row shown.
  var selection = document.getElementById('selection');
  if (selection) {
    var rows = function () {
      return document.querySelectorAll('input[name="ids"][form="selection"]');
    };
    selection.addEventListener('submit', function (event) {
      var picked = Array.prototype.some.call(rows(), function (box) { return box.checked; });
      if (!picked) {
        event.preventDefault();
        window.alert(selection.dataset.empty);
      }
    });
    var all = document.getElementById('select-all');
    if (all) {
      all.addEventListener('change', function () {
        rows().forEach(function (box) { box.checked = all.checked; });
      });
    }
  }

  // The create token form: the date field shows only for a custom expiration.
  var expiration = document.getElementById('expiration');
  var group = document.getElementById('date-group');
  var date = document.getElementById('date');
  if (expiration && group && date) {
    var sync = function () {
      var custom = expiration.value === 'custom';
      group.hidden = !custom;
      date.required = custom;
    };
    expiration.addEventListener('change', sync);
    sync();
  }

  // The create token form: add the browser's own ip to the allowed ones.
  var mine = document.getElementById('use-my-ip');
  if (mine) {
    mine.addEventListener('click', function () {
      var field = document.getElementById('ips');
      var ip = mine.dataset.ip;
      var listed = field.value.split(',').map(function (entry) { return entry.trim(); }).filter(Boolean);
      if (listed.indexOf(ip) === -1) {
        listed.push(ip);
      }
      field.value = listed.join(', ');
    });
  }

  // The visualization settings form: add the browser's own ip to the ignored
  // ones, one per line.
  var ignore = document.getElementById('visualization-add-my-ip');
  if (ignore) {
    ignore.addEventListener('click', function () {
      var field = document.getElementById('ignored-ips');
      var ip = ignore.dataset.ip;
      var listed = field.value.split('\n').map(function (entry) { return entry.trim(); }).filter(Boolean);
      if (listed.indexOf(ip) === -1) {
        listed.push(ip);
      }
      field.value = listed.join('\n');
    });
  }

  // The snapshot upload form: the server reads no multipart body, so the
  // file is sent as the whole body of the request, and the list is reloaded
  // with the outcome — or the reason it was refused is shown.
  var upload = document.getElementById('snapshot-upload');
  if (upload) {
    upload.addEventListener('submit', function (event) {
      event.preventDefault();
      var file = document.getElementById('snapshot-file').files[0];
      var button = document.getElementById('snapshot-upload-button');
      var error = document.getElementById('upload-error');
      if (!file) {
        return;
      }
      var fail = function (message) {
        error.textContent = message;
        error.hidden = false;
        button.disabled = false;
        button.textContent = label;
      };
      var label = button.textContent;
      error.hidden = true;
      button.disabled = true;
      button.textContent = upload.dataset.uploading;
      fetch(upload.action, {
        method: 'POST',
        headers: { 'Content-Type': 'application/zip' },
        body: file
      }).then(function (response) {
        if (response.ok) {
          location.href = '/admin/root/list-backups?notice=uploaded';
          return;
        }
        return response.json().then(function (body) {
          fail(body.error || upload.dataset.refused);
        }, function () {
          fail(upload.dataset.refusedStatus.replace('{0}', response.status));
        });
      }, function () {
        fail(upload.dataset.unreachable);
      });
    });
  }

  // The token just created: select it on focus, copy it, and spell the
  // example request against this server.
  var origin = document.getElementById('origin');
  if (origin) {
    origin.textContent = location.origin;
  }
  var token = document.getElementById('new-token');
  if (token) {
    token.addEventListener('focus', function () { token.select(); });
  }
  var copy = document.getElementById('copy-token');
  if (copy && token) {
    copy.addEventListener('click', function () {
      token.select();
      var done = function () { copy.textContent = copy.dataset.copied; };
      if (navigator.clipboard) {
        navigator.clipboard.writeText(token.value).then(done, function () { document.execCommand('copy'); done(); });
      } else {
        document.execCommand('copy');
        done();
      }
    });
  }
})();
