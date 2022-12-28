#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Des 6 2022
@author: Hue (MohammadHossein) Salari
@email:hue.salari@gmail.com

Sources:
    - https://towardsdatascience.com/how-to-extract-data-from-the-twitter-api-using-python-b6fbd7129a33
"""


import os
import logging as log
from tqdm import tqdm
from dotenv import load_dotenv
from datetime import datetime, timedelta

import tweepy
import requests
from urlextract import URLExtract

# import helper functions from other modules
from utils.send_email import send_email
from utils.compose_email import compose_email
from utils.phd_keywords import phd_keywords
from utils.customers_data import customers_data

# set up logging to a file
log_file_path = os.path.join(
    os.path.dirname(os.path.realpath(__file__)), "temp", "twitter.log"
)
log.basicConfig(
    level=log.INFO,
    filename=log_file_path,
    format="%(asctime)s %(levelname)s %(message)s",
)


import tweepy
import os


def connect_to_twitter_api():
    """
    Connect to the Twitter API.

    Returns:
        tweepy.API: Twitter API client.
    """

    # load environment variables from .env file
    load_dotenv()

    # get API keys and access tokens from environment variables
    consumer_key = os.environ["API_KEY"]
    consumer_secret = os.environ["API_KEY_SECRET"]
    access_token = os.environ["ACCESS_TOKEN"]
    access_token_secret = os.environ["ACCESS_TOKEN_SECRET"]

    # authenticate with the Twitter API
    auth = tweepy.OAuth1UserHandler(
        consumer_key, consumer_secret, access_token, access_token_secret
    )

    # create a Twitter API client
    api = tweepy.API(auth)

    return api


def extract_images_url_in_tweet_status(tweet):
    """
    Extracts image URLs from a tweet object.

    Parameters:
    tweet (dict): A tweet object from the Twitter API.

    Returns:
    list: A list of image URLs.
    """
    # Check if the tweet has media entities
    if "media" in tweet.entities:
        # Extract image URLs from the tweet's media entities
        urls = [x["media_url"] for x in tweet.entities["media"]]

        # Check if the tweet has extended media entities
        try:
            # Extract image URLs from the tweet's extended media entities
            extra = [x["media_url"] for x in tweet.extended_entities["media"]]

            # Combine the lists of image URLs
            urls = set(urls + extra)
        except:
            # If there are no extended media entities, do nothing
            pass

        # Return the list of image URLs
        return urls
    else:
        # If there are no media entities, return None
        return None


def find_positions(twitter_api, phd_keywords, date=None, since_id=None):
    """
    Find tweets containing keywords related to  funded PhD positions.

    Parameters:
        twitter_api (tweepy.API): Twitter API client.
        phd_keywords (List[str]): List of keywords to search for in Twitter.
        date (str): Date in the format "YYYY-MM-DD" to use as a lower bound for the the search.

    Returns:
        tweets set[tweepy.Status]: set of tweets containing the given keywords that posted after the given date.
    """

    # list to store found tweets
    tweets = list()

    print("Searching for PhD positions in Twitter🐤")
    # iterate through PhD keywords and search for tweets containing each keyword
    pbar = tqdm(phd_keywords)

    for keyword in pbar:
        pbar.set_postfix(
            {
                "Keyword": keyword,
                "TN of founded positions": len(tweets),
            }
        )
        # create query string for searching tweets containing the keyword and search Twitter for tweets matching the query

        if since_id:  # posted after the given tweet_id
            query = f'"{keyword}" -filter:retweets'
            found_tweets = twitter_api.search_tweets(
                query,
                tweet_mode="extended",
                result_type="recent",
                lang="en",
                count=100,
                since_id=since_id,
            )
        elif date:  # posted after the given date
            query = f'"{keyword}" since:{date} -filter:retweets'
            found_tweets = twitter_api.search_tweets(
                query, tweet_mode="extended", result_type="recent", lang="en", count=100
            )
        else:
            query = f'"{keyword}" -filter:retweets'
            found_tweets = twitter_api.search_tweets(
                query, tweet_mode="extended", result_type="recent", lang="en", count=100
            )
        # add found tweets to the list of tweets
        tweets += found_tweets
        tweets = list(set(tweets))
    return tweets


def clean_tweets(tweets):
    """
    Filter and format tweets.

    Parameters:
        tweets (List[tweepy.Status]): List of tweets to filter and format.

    Returns:
        positions, raw_positions (List[str], List[tweepy.Status]): List of formatted tweets and raw tweets
    """

    # list to store formatted tweets
    positions = []

    # list to store raw tweets
    raw_positions = []

    # list to store previously seen tweet text
    seen_positions = []

    # iterate through tweets
    for tweet in set(tweets):
        url = ""
        urls = []
        try:
            # get tweet description from retweeted status if possible
            text = tweet.retweeted_status.full_text
        except AttributeError:
            # get tweet description from original tweet if retweeted status is not available
            text = tweet.full_text
        # format tweet if its text has not been seen before
        if text not in seen_positions:
            # add tweet description to the list of seen descriptions
            seen_positions.append(f"{text}")

            # fix &amp; problem
            text = text.replace("&amp;", "&")

            # Replace shortened URLs with distention URL
            extractor = URLExtract()
            urls = extractor.find_urls(text)
            for _url in urls:
                try:
                    response = requests.get(_url, allow_redirects=True)
                    final_url = response.url
                    text = text.replace(_url, final_url)
                except:
                    pass

            # generate URL for the tweet
            url = f"https://twitter.com/twitter/statuses/{tweet.id}"

            # format tweet and add it to the list of formatted tweets
            formatted_tweet = f"date: {tweet.created_at.strftime('%Y-%m-%d %H:%M:%S')}"
            formatted_tweet += f"<br>by: {tweet.user.name}"
            formatted_tweet += f"<br><br>{text}"
            formatted_tweet += f"<br><br>🐦🔗:<br>{url}"
            positions.append(formatted_tweet)

            # Add the tweet to the list of raw tweets
            raw_positions.append(tweet)
    return positions, raw_positions


def filter_positions(positions, keywords):
    """
    Filter list of positions based on keywords.

    Parameters:
        positions (List[str]): List of positions to filter.
        keywords (List[str]): List of keywords to use for filtering.

    Returns:
        List[str]: Filtered list of positions.
    """

    # list to store filtered positions
    results = []
    # iterate through positions
    for text in positions:
        # include the position in results if it contains any of the keywords
        if any(item.lower() in text.lower() for item in keywords):
            results.append(text)
    return results


def compose_and_send_email(recipient_email, recipient_name, positions, base_path):
    """
    Compose and send email containing positions.

    Parameters:
        recipient_email (str): Email address to send the email to.
        recipient_name (str): Customer's name to include in the email.
        positions (List[str]): List of positions to include in the email.
        base_path (str): Base path for the email template file.

    Returns:
        None
    """
    # generate email text using the email template and the given positions
    email_text = compose_email(recipient_name, "Twitter", positions, base_path)
    # send email with the generated text
    send_email(recipient_email, "PhD Positions from Twitter", email_text, "html")


def main():
    # Log message
    log.info(f"Searching Twitter for Ph.D. Positions")

    # Connect to Twitter API
    twitter_api = connect_to_twitter_api()

    # Set base path
    base_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "utils")

    # Get date from yesterday
    yesterday = datetime.today() - timedelta(days=1)
    date = yesterday.strftime("%Y-%m-%d")

    # Search for Ph.D. positions and clean up tweets
    phd_positions, _ = clean_tweets(find_positions(twitter_api, phd_keywords[:], date))

    # Log number of tweets that have been found
    log.info(f"Found {len(phd_positions)} tweets related to Ph.D. Positions")

    # Loop through customers
    for customer in customers_data:
        # Check if customer's expiration date is today or later
        if customer["expiration_date"] >= datetime.today():
            # Get keywords for customer
            customer_keywords = set(
                [keyword.replace(" ", "") for keyword in customer["keywords"]]
                + customer["keywords"]
            )

            # Filter positions based on customer keywords
            related_positions = filter_positions(phd_positions, customer_keywords)

            # If there are related positions, send email to customer
            if related_positions:
                log.info(
                    f"Sending email containing {len(related_positions)} positions to: {customer['name']}"
                )
                tqdm.write(f"sending email to: {customer['name']}")
                compose_and_send_email(
                    customer["email"], customer["name"], related_positions, base_path
                )


if __name__ == "__main__":
    main()
