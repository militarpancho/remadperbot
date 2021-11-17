#!/usr/bin/env python
# coding: utf-8

import requests
import re
from datetime import datetime
from bs4 import BeautifulSoup
import io
import telegram
import time
import logging
import os

LOGLEVEL = os.environ.get('LOGLEVEL', 'INFO').upper()
TOKEN=os.environ.get("TOKEN")
CHANNEL_ID=os.environ.get("CHANNEL_ID")
SCRAPE_INTERVAL=os.environ.get("SCRAPE_INTERVAL", 10)
SCRAPE_WINDOW=os.environ.get("SCRAPE_WINDOW", 100)

logging.basicConfig(level=LOGLEVEL, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')

headers = {"User-Agent": ""}

def extract_article_info(raw_html_data, url):
    soup = BeautifulSoup(raw_html_data, 'html.parser')
    title = soup.h1
    image = soup.find_all("img")[2]
    data = list(map(lambda x: clean_category(x.contents),title.find_all_next("p")[0:5]))
    title = "<a href=\"{}\">{}</a>".format(url,title.text)
    img = image['src']
    return { "metadata": data, "title": title, "img": img }

def clean_category(x):
    regex = re.compile(r'[\n\r\t]')
    cleaned = ""
    for i in x:
        cleaned += regex.sub("", str(i)) + " "
    return cleaned

def check_if_green_point_opened():
    now = datetime.now()
    today8am = now.replace(hour=8, minute=0, second=0, microsecond=0)
    today8pm = now.replace(hour=20, minute=0, second=0, microsecond=0)
    return today8am < now < today8pm

def main():
    bot = telegram.Bot(token=TOKEN)
    current_id = int(os.environ.get("FIRST_ID", 31158))
    while True:
        url = "https://www.remad.es/web/antiquity/{}".format(current_id)
        try:
            r = requests.get(url, verify=False, headers=headers)
            logging.debug("Request to {} with status code {}".format(r.url, r.status_code))
            if r.status_code == 200:
                data = extract_article_info(r.content, url)
                current_id +=1
                logging.info("New Product found, increasing counter to {}".format(current_id))
                r = requests.get(data['img'].replace("./",""), stream=True, verify=False, allow_redirects=True, headers=headers)
                img = io.BytesIO(r.content)
                try:
                    bot.send_photo(chat_id=CHANNEL_ID, photo=img, caption=data['title'] + '\n' + "\n".join(data['metadata']), parse_mode=telegram.ParseMode.HTML)
                except:
                    logging.debug("Image of {} can't be sent".format(current_id))
            else:
                if check_if_green_point_opened():
                    time.sleep(int(SCRAPE_INTERVAL))
                    logging.debug("Green points open, finding more products...")
                    for i in range(1, int(SCRAPE_WINDOW)):
                        url = "https://www.remad.es/web/antiquity/{}".format(current_id+i)
                        try: 
                            r = requests.get(url, verify=False, headers=headers)
                            if r.status_code == 200:
                                current_id = current_id+i
                                break
                        except:
                            logging.info("Request not processable")
                        time.sleep(0.1)
                else:
                    time.sleep(600)
        except:
            logging.info("Error getting data from request. Continue to the next product")
            current_id += 1 


if __name__ == '__main__':
    main()

