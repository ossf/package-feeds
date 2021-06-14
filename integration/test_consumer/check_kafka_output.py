from datetime import datetime
from confluent_kafka import Consumer

import time
import requests

def trigger_feeds():
    attempts = 5
    print("Requesting feeds poll data from registries...")
    while True:
        try:
            r = requests.get('http://127.0.0.1:8080')
            break
        except:
            if attempts == 0:
                raise
            print("Warning: Failed to request data from package-feeds, retrying after 5s")
            time.sleep(5)

        attempts -= 1

    print(r.text)



def main():
    msgs = []

    trigger_feeds()

    c = Consumer({
        'bootstrap.servers': '127.0.0.1:9094',
        'group.id': 'consumer',
        'auto.offset.reset': 'earliest'
    })

    c.subscribe(['package-feeds'])

    last_poll_success = datetime.now()
    while True:
        msg = c.poll(2.0)

        if msg is None:
            delta = datetime.now() - last_poll_success
            # Timeout to avoid hanging on poll() loop
            if delta.total_seconds() > 10:
                break
            continue
        if msg.error():
            print("Consumer error: {}".format(msg.error()))
            continue
        msgs.append(msg)
        last_poll_success = datetime.now()
        print('Received message: {}'.format(msg.value().decode('utf-8')))

    print(f"\n\n------------------------------------------------------------\n\nReceived a total of {len(msgs)} messages")
    c.close()
    assert len(msgs) > 0, "Failed to assert that atleast a single package was received"

if __name__ == '__main__':
    main()