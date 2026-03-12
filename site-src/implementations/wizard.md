# Controller Matching Wizard

<div id="wizard-iframe-container">
  <iframe id="wizard-iframe" src="../../wizard/" title="Controller matching wizard" style="width: 100%; min-height: 400px; border: none; display: block;"></iframe>
</div>

<script>
(function() {
  var iframe = document.getElementById('wizard-iframe');
  if (!iframe) return;
  function setHeight(h) {
    iframe.style.height = (typeof h === 'number' ? h : 400) + 'px';
  }
  window.addEventListener('message', function(event) {
    if (event.data && event.data.type === 'wizard-height' && typeof event.data.height === 'number') {
      setHeight(event.data.height);
    }
  });
  setHeight(400);
})();
</script>
