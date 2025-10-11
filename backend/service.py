import logging
import time
from http import HTTPStatus

from fastapi import FastAPI, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
import uvicorn

from backend.core.config.env import load_environment
from backend.core.obs.logging import setup_logger
from backend.core.utils.types import JsonResponse


logger = logging.getLogger(__name__)
config = load_environment()
app = FastAPI(
    title=config.app_name,
    version=config.app_version,
    description="Backend service for software component management",
    contact={
        "name": "Chino Franco",
        "email": "chino.franco@gmail.com",
        "github": "https://github.com/jgfranco17"
    },
)
startup_time = time.time()


@app.get("/", status_code=HTTPStatus.OK, tags=["SYSTEM"])
def root():
    """Project main page."""
    return {
        "message": f"Welcome to the {config.app_name}!",
        "version": config.app_version,
        "debug": config.debug
    }


@app.get("/healthz", status_code=HTTPStatus.OK, tags=["SYSTEM"])
def health_check() -> JsonResponse:
    """Health check for the API."""
    return {"status": "ok", "uptime": time.time() - startup_time}


@app.get("/config", status_code=HTTPStatus.OK, tags=["SYSTEM"])
def get_config() -> JsonResponse:
    """Get the current application configuration."""
    return {
        "app_name": config.app_name,
        "app_version": config.app_version,
        "debug": config.debug,
        "host": config.host,
        "port": config.port,
        "log_level": config.log_level,
    }


@app.exception_handler(HTTPException)
async def http_exception_handler(request: Request, exc: HTTPException):
    """General exception handler."""
    return JSONResponse(
        status_code=exc.status_code,
        content={
            "message": exc.detail,
            "request": {
                "method": request.method,
                "url": str(request.url),
                "status": exc.status_code,
            },
        },
    )


app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


def start():
    """Main entry point for the application."""
    setup_logger()
    logger.info(f"Starting {config.app_name} v{config.app_version}...")
    uvicorn.run(
        app,
        host=config.host,
        port=config.port,
        reload=config.reload,
        log_level=config.log_level.lower()
    )