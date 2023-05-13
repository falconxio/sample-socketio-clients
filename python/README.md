## HOW TO RUN

1. Make sure you have Python version 3.9.16 or higher and python-socketio version 4.6.0
2. Confirm installation version
```
python --version
pip show python-socketio
```
3. cd edison/clients/sample-socketio-clients/python/socketioclient
4. Replace API Key/Pass Phrase/Secret in socketio_client.py
5. Edit "--token_pairs" in ArgumentParser in socketio_client.py to tokenPairs you want to test for
6. Similarly edit "--levels" in ArgumentParser in socketio_client.py
7. run "python socketio_client.py"
8. Verify the subscription successful response.
