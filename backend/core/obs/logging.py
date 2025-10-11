import logging
import os
from typing import Final

from backend.core.config.env import LOG_LEVEL

logger = logging.getLogger(__name__)


LOG_FORMAT: Final[str] = "[%(asctime)s][%(levelname)s] %(name)s: %(message)s"
TIMESTAMP_FORMAT: Final[str] = "%Y-%m-%d %H:%M:%S"


def setup_logger():
    """Set up logging configuration."""
    levels = {
        "CRITICAL": logging.CRITICAL,
        "ERROR": logging.ERROR,
        "WARNING": logging.WARNING,
        "INFO": logging.INFO,
        "DEBUG": logging.DEBUG,
    }
    base_log_level = logging.INFO
    if level_from_env := os.getenv(LOG_LEVEL):
        base_log_level = levels.get(level_from_env.upper(), base_log_level)
    logging.basicConfig(
        format=LOG_FORMAT, datefmt=TIMESTAMP_FORMAT, level=base_log_level
    )
    logger.debug(
        f"Logging initialized with level: {logging.getLevelName(base_log_level)}"
    )
