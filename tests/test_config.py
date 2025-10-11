"""Test the configuration model."""

import os
import tempfile
from unittest.mock import patch

import pytest
from pytest import MonkeyPatch
from tomlkit import value

from backend.core.config.env import Config, load_environment


def test_config_defaults():
    """Test that the configuration has sensible defaults."""
    default_config = Config()

    assert default_config.app_name == "DevOps API"
    assert default_config.app_version == "0.0.1"
    assert default_config.debug is False
    assert default_config.host == "0.0.0.0"
    assert default_config.port == 8000
    assert default_config.reload is False
    assert default_config.log_level == "INFO"


def test_config_from_env(monkeypatch: MonkeyPatch):
    """Test that configuration can be loaded from environment variables."""
    env_vars = {
        "APP_NAME": "Test API",
        "DEBUG": "true",
        "PORT": "3000",
        "LOG_LEVEL": "DEBUG",
        "CORS_ORIGINS": "http://localhost:3000,https://example.com",
    }
    for key, value in env_vars.items():
        monkeypatch.setenv(key, value)

    test_config = Config.from_env()

    assert test_config.app_name == "Test API"
    assert test_config.debug is True
    assert test_config.port == 3000
    assert test_config.log_level == "DEBUG"


@pytest.mark.parametrize("value", ["true", "True", "TRUE", "1", "t", "yes", "y", "on"])
def test_config_bool_parsing_true(monkeypatch: MonkeyPatch, value: str):
    """Test boolean parsing from environment variables (truthy)."""
    monkeypatch.setenv("DEBUG", value)
    test_config = Config.from_env()
    assert test_config.debug is True, f"Value '{value}' evaluated to False"


@pytest.mark.parametrize(
    "value", ["false", "False", "FALSE", "0", "f", "no", "n", "off", ""]
)
def test_config_bool_parsing_false(monkeypatch: MonkeyPatch, value: str):
    """Test boolean parsing from environment variables (falsy)."""
    monkeypatch.setenv("DEBUG", value)
    test_config = Config.from_env()
    assert test_config.debug is False, f"Value '{value}' evaluated to True"
