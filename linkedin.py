#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Des 6 2022
@author: Hue (MohammadHossein) Salari
@email:hue.salari@gmail.com

Sources: 
    - https://www.geeksforgeeks.org/scrape-linkedin-using-selenium-and-beautiful-soup-in-python/
    - https://stackoverflow.com/questions/64717302/deprecationwarning-executable-path-has-been-deprecated-selenium-python
    - https://stackoverflow.com/questions/32391303/how-to-scroll-to-the-end-of-the-page-using-selenium-in-python
"""


import os
import time
from tqdm import tqdm
from dotenv import load_dotenv
from datetime import datetime, timedelta

from bs4 import BeautifulSoup

from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from webdriver_manager.chrome import ChromeDriverManager

from utils.send_email import send_email
from utils.compose_email import compose_email

from utils.phd_keywords import phd_keywords
from utils.customers_data import customers_data


def make_driver():
    options = Options()
    # options.add_argument("--headless")
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")

    driver = webdriver.Chrome(
        service=Service(ChromeDriverManager().install()), options=options
    )

    return driver


def login_to_linkedin(driver):
    email = os.environ["LINKEDIN_EMAIL_ADDRESS"]
    password = os.environ["LINKEDIN_PASSWORD"]

    driver.get("https://linkedin.com/uas/login")
    time.sleep(5)
    username = driver.find_element("id", "username")
    username.send_keys(email)
    pword = driver.find_element("id", "password")
    pword.send_keys(password)
    driver.find_element("xpath", "//button[@type='submit']").click()


def find_positions(driver, phd_keywords):
    _positions = []
    positions = []
    for keyword in tqdm(phd_keywords):
        url = f'https://www.linkedin.com/search/results/content/?datePosted=%22past-24h%22&keywords="{keyword}"&origin=FACETED_SEARCH&sid=c%3Bi&sortBy=%22date_posted%22'
        driver.get(url)
        for _ in tqdm(range(10), leave=False):
            driver.execute_script("window.scrollTo(0, document.body.scrollHeight);")
            time.sleep(6)

        src = driver.page_source
        soup = BeautifulSoup(src, "lxml")

        search_results = soup.find("main", {"aria-label": "Search results"})

        _positions += search_results.find_all(
            "div",
            {
                "class": "update-components-text relative feed-shared-update-v2__commentary"
            },
        )
        positions += [position.text.strip() for position in _positions]

        links_p1 = search_results.find_all(
            "div",
            {
                "class": "feed-shared-update-v2 feed-shared-update-v2--minimal-padding full-height relative feed-shared-update-v2--e2e artdeco-card"
            },
        )

        links_p2 = search_results.find_all(
            "div",
            {
                "class": "feed-shared-update-v2 feed-shared-update-v2--minimal-padding full-height relative artdeco-card"
            },
        )
        for links in links_p1 + links_p2:
            try:
                position_text = links.find(
                    "div",
                    {
                        "class": "update-components-text relative feed-shared-update-v2__commentary"
                    },
                ).text.strip()
                if position_text:
                    positions[
                        positions.index(position_text)
                    ] = f'{position_text}\nhttps://www.linkedin.com/feed/update/{links["data-urn"]}'
            except:
                pass

    return list(set(positions))


def filter_positions(positions, keywords):
    results = []
    for position in positions:
        if any(item.lower() in position.lower() for item in keywords):
            results.append(position)
    return results


def compose_and_send_email(email_address, customers_name, positions, base_path):

    email_text = compose_email(customers_name, "LinkedIn", positions, base_path)

    send_email(email_address, "PhD Positions from LinkedIn", email_text, "html")


if __name__ == "__main__":

    base_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "utils")

    load_dotenv()
    print("Opening Chrome")
    driver = make_driver()
    print("Logging in to the Linkedin")
    login_to_linkedin(driver)
    print("Searching for the fully funded Ph.D. positions in Linkedin 🐷")
    positions = find_positions(driver, phd_keywords[:])
    driver.quit()
    print(f"Total number of positions: {len(positions)}")

    today = datetime.today()
    date = today - timedelta(days=1)
    date = date.strftime("%Y-%m-%d")

    for customer in customers_data:
        if customer["expiration_date"] >= today:
            keywords = set(
                [keyword.replace(" ", "") for keyword in customer["keywords"]]
                + customer["keywords"]
            )
            related_positions = filter_positions(positions, keywords)
            if related_positions:
                print(f"sending email to: {customer['name']}")
                compose_and_send_email(
                    customer["email"], customer["name"], related_positions, base_path
                )
                time.sleep(10)
