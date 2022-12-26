import os
import wget
import pickle
import logging as log
from tqdm import tqdm
from dotenv import load_dotenv
from datetime import datetime, timedelta

import twitter
from mastodon import Mastodon

from utils.phd_keywords import phd_keywords

# set up logging to a file
log_file_path = os.path.join(os.path.dirname(os.path.realpath(__file__)), "twitter.log")
log.basicConfig(
    level=log.INFO,
    filename=log_file_path,
    format="%(asctime)s %(levelname)s %(message)s",
)
import requests

keywords = [
    "computer vision",
    "deep learning",
    "machine learning",
    "artificial intelligence",
    "remote sensing",
    "object detection",
    "classification",
    "segmentation",
    "object extraction",
    "supervised learning",
    "semi-supervised learning",
    "unsupervised learning",
    "weakly-supervised learning",
    "#AI",
    "#ML",
    "NLP",
    "data science" "natural Language processing.",
]
keywords = set([keyword.replace(" ", "") for keyword in keywords] + keywords)


def set_up_mastodon():
    """
    Sets up a Mastodon API client.

    Returns:
    Mastodon: A Mastodon API client.
    """

    load_dotenv()
    access_token = os.environ["MASTODON_ACCESS_TOKEN"]

    # Create a Mastodon API client using the provided access token
    mastodon = Mastodon(
        access_token=access_token,
        api_base_url="https://sigmoid.social/",
    )

    # Return the Mastodon API client
    return mastodon


def download_media_and_post_to_mastodon(mastodon_api, text, images_url, temp_path):
    # Download tweets media if available and upload to Mastodon
    media_ids = []
    if images_url:
        images_filename = []
        for image_url in tqdm(images_url):
            # Download the image
            try:
                file_name = wget.download(image_url, out=temp_path)
                images_filename.append(file_name)
            except Exception as e:
                print(e)
                continue
        # Upload the images to Mastodon and remove uploaded file from disk
        for image in images_filename:
            media_ids.append(mastodon_api.media_post(image))
            os.remove(image)
    # send the toot to Mastodon
    mastodon_api.status_post(text, media_ids=media_ids)


def main():
    # Log message
    log.info(f"Searching Twitter for Ph.D. Positions")

    # Connect to Twitter API
    twitter_api = twitter.connect_to_twitter_api()

    # Connect to Mastodon API
    mastodon_api = set_up_mastodon()

    # Set tmp path and and create it if not exists
    temp_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "temp")
    if not os.path.exists(temp_path):
        os.mkdir(temp_path)

    newest_tweet_id_path = os.path.join(
        os.path.dirname(os.path.abspath(__file__)),
        "temp",
        "mastodon_twitter_search_newest_tweet_id.pickle",
    )

    # Search for Ph.D. positions
    # if it's the first time we are running this code, search for positions it the past 7 days.
    # otherwise, search seance the id of last tweet
    if os.path.exists(newest_tweet_id_path):
        since_id = pickle.load(open(newest_tweet_id_path, "rb"))
        tweets = twitter.find_positions(twitter_api, phd_keywords[:], since_id=since_id)
    else:
        # Get date from yesterday
        yesterday = datetime.today() - timedelta(days=7)
        date = yesterday.strftime("%Y-%m-%d")
        tweets = twitter.find_positions(twitter_api, phd_keywords[:], date)

    # Save the id of newest tweet
    ids = list()
    for tweet in tweets:
        ids.append(tweet.id)
    ids.sort()
    pickle.dump(ids[-1], open(newest_tweet_id_path, "wb"))

    # Extract postilions test and images from tweets
    positions_text = twitter.clean_tweets(tweets)
    positions_images = [
        twitter.extract_images_url_in_tweet_status(tweet) for tweet in tweets
    ]

    for text, images_url in zip(positions_text, positions_images):
        if any(item.lower() in text.lower() for item in keywords):
            # Format the text of position
            text = text.replace("<br>", "\n")
            text += "\n #PhdPosition"
            download_media_and_post_to_mastodon(
                mastodon_api, text, images_url, temp_path
            )

    # # Log number of tweets that have been found
    # log.info(f"Found {len(tweets)} tweets related to Ph.D. Positions")


if __name__ == "__main__":
    main()
