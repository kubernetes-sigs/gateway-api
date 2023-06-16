import shutil
import logging
from mkdocs import plugins

log = logging.getLogger('mkdocs')

@plugins.event_priority(100)
def on_pre_build(config, **kwargs):
    log.info("copying geps")
    shutil.copytree("geps","site-src/geps", dirs_exist_ok=True)
