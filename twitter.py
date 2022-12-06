# https://towardsdatascience.com/how-to-extract-data-from-the-twitter-api-using-python-b6fbd7129a33

import os
from tqdm import tqdm
from dotenv import load_dotenv
from datetime import datetime, timedelta

import tweepy

from utils.send_email import send_email
from utils.compose_email import compose_email
from utils.phd_keywords import phd_keywords
from utils.customers_data import customers_data


def connect_to_twitter_api():
    load_dotenv()

    consumer_key = os.environ["API_KEY"]
    consumer_secret = os.environ["API_KEY_SECRET"]
    access_token = os.environ["ACCESS_TOKEN"]
    access_token_secret = os.environ["ACCESS_TOKEN_SECRET"]

    auth = tweepy.OAuth1UserHandler(
        consumer_key, consumer_secret, access_token, access_token_secret
    )

    api = tweepy.API(auth)

    return api


def find_positions(twitter_api, keywords, date):

    tweets = []
    print("Searching for full funded phd positions in Twitter🐤")
    for keyword in tqdm(keywords):
        query = f'"{keyword}" since:{date} -filter:retweets'
        tweets += twitter_api.search_tweets(
            query, tweet_mode="extended", result_type="recent", lang="en", count=100
        )

    return tweets


def clean_tweets(tweets):
    positions = []
    _positions = []
    for tweet in set(tweets):
        url = f"https://twitter.com/twitter/statuses/{tweet.id}"
        try:
            description = tweet.retweeted_status.full_text

        except AttributeError:
            description = tweet.full_text

        if description not in _positions:
            _positions.append(f"{description}")
            positions.append(
                f"date:{tweet.created_at.strftime('%Y-%m-%d %H:%M:%S')}\nby:{tweet.user.name}\n{description}\n{url}"
            )
    return positions


def filter_positions(positions, keywords):
    results = []
    for text in positions:
        if any(item.lower() in text.lower() for item in keywords):
            results.append(text)
    return results


def compose_and_send_email(email_address, positions):

    email_text = compose_email(positions)
    send_email(email_address, "PhD Positions from Twitter", email_text, "html")


if __name__ == "__main__":

    api = connect_to_twitter_api()

    today = datetime.today()
    date = today - timedelta(days=1)
    date = date.strftime("%Y-%m-%d")

    tweets = clean_tweets(find_positions(api, phd_keywords, date))

    for customer in customers_data:
        if customer["expiration_date"] >= today:
            keywords = set(
                [keyword.replace(" ", "") for keyword in customer["keywords"]]
                + customer["keywords"]
            )
            related_positions = filter_positions(tweets, keywords)
            if related_positions:
                tqdm.write(f"sending email to: {customer['name']}")
                compose_and_send_email(customer["email"], related_positions)
