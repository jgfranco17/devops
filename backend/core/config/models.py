import yaml
from pathlib import Path

from pydantic import BaseModel


class SoftwareComponent(BaseModel):
    """Model representing a software component with name and version."""

    name: str
    version: str
    description: str | None = None
    repository: str | None = None

    @classmethod
    def from_definition_file(cls, filepath: Path) -> "SoftwareComponent":
        """Load a Software Component from a definition file."""
        with open(filepath, "r") as f:
            if filepath.suffix.lower() not in (".yml", ".yaml"):
                raise ValueError("Unsupported file format. Use .json or .yaml/.yml")
            data = yaml.safe_load(f)
        return cls(**data)