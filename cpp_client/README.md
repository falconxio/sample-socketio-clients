## HOW TO RUN

1. Run the install_dependencies.sh bash file to install the necessary dependencies (only for ubuntu)
2. cd edison/clients/cpp_client
3. Replace API Key/Pass Phrase/Secret in client.cpp
4. Build the file using the following command

   ```
   g++ client.cpp -o main -lsioclient_tls -lboost_system -lssl -lcrypto -pthread -std=c++11
   ```
5. Run the built file using the below command, to subscribe to ETH/USD and BTC/USD markets for 1,5,10 levels each.

   ```
   ./main ETH/USD,BTC/USD 1,5,10
   ```
