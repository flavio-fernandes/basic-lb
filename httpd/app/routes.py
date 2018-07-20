from app import app
import os

@app.route('/')
def index():
    return os.environ.get("MSG", "Hello, World!")
