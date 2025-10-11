import logging
import os

from pydantic import BaseModel, Field

logger = logging.getLogger(__name__)

# Constants for backward compatibility
LOG_LEVEL = "LOG_LEVEL"


class Config(BaseModel):
    """Application configuration loaded from environment variables."""

    # Application settings
    app_name: str = Field(default="DevOps API")
    app_version: str = Field(default="0.0.1")
    debug: bool = Field(default=False)

    # Server settings
    host: str = Field(default="0.0.0.0")
    port: int = Field(default=8000)
    reload: bool = Field(default=False)

    # Logging settings
    log_level: str = Field(default="INFO")

    @classmethod
    def from_env(cls) -> "Config":
        """Create a Config instance from environment variables."""

        def parse_bool(value: str) -> bool:
            """Parse a string as a boolean."""
            return value.lower() in ("true", "1", "t", "yes", "y", "on")

        def parse_list(value: str) -> list[str]:
            """Parse a comma-separated string as a list."""
            if not value:
                return []
            return [item.strip() for item in value.split(",")]

        def parse_log_level(value: str) -> str:
            """Parse and validate log level."""
            level = value.upper()
            valid_levels = {"DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"}
            if level not in valid_levels:
                return "INFO"
            return level  # type: ignore

        return cls(
            app_name=os.getenv("APP_NAME", "DevOps API"),
            app_version=os.getenv("APP_VERSION", "0.0.1"),
            debug=parse_bool(os.getenv("DEBUG", "false")),
            host=os.getenv("HOST", "0.0.0.0"),
            port=int(os.getenv("PORT", "8000")),
            reload=parse_bool(os.getenv("RELOAD", "false")),
            log_level=parse_log_level(os.getenv("LOG_LEVEL", "INFO")),
        )


def load_environment() -> Config:
    """Load the application configuration from environment variables."""
    config = Config.from_env()
    logger.info("Environment configuration loaded successfully!")
    return config
