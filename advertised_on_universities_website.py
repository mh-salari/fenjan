import os

from utils.search import search_for_keywords
from utils.compose_email import compose_and_send_email
from utils.keep_track_of_sent_emails import TrackingEmailsDatabase, TrackingEmails
from utils.database_helpers import *
from utils.universities import universities

from tqdm import tqdm

# set up logging to a file
import logging as log

log_file_path = os.path.join(
    os.path.dirname(os.path.realpath(__file__)),
    "temp",
    "kth_royal_institute_of_technology.log",
)
log.basicConfig(
    level=log.INFO,
    filename=log_file_path,
    format="%(asctime)s %(levelname)s %(message)s",
)

log.getLogger().addHandler(log.StreamHandler())


def filter_and_send_emails(uni):

    # Set base path
    base_path = os.path.dirname(os.path.abspath(__file__))

    utils_path = os.path.join(base_path, "utils")
    dotenv_path = os.path.join(base_path, ".env")

    log.info("Getting sent_emails info.")
    tracking_emails_db = TrackingEmailsDatabase(*get_db_connection_values(dotenv_path))
    tracking_emails_cnx = tracking_emails_db.connect_to_database()
    log.info("Getting customers info.")
    customers = get_customers_info(dotenv_path)
    log.info(f"Getting {uni.name} positions info.")

    positions = getting_positions_info(dotenv_path, uni)

    for customer in tqdm(customers):
        customer_keywords = list(
            set(
                [keyword.replace(" ", "") for keyword in customer.keywords]
                + customer.keywords
            )
        )

        log.info(f"Filtering Positions based of {customer.email} keywords")
        related_positions = []
        for position in positions:

            customer_search_result = search_for_keywords(
                position.title + position.description, keywords=customer_keywords
            )

            if uni.target_keywords != []:

                target_keywords_search_result = search_for_keywords(
                    position.title, uni.target_keywords
                )
            else:
                target_keywords_search_result = True

            if uni.forbidden_keywords != []:

                forbidden_words_search_result = search_for_keywords(
                    position.title, uni.forbidden_keywords
                )
            else:
                forbidden_words_search_result = False

            if (
                customer_search_result
                and target_keywords_search_result
                and not forbidden_words_search_result
            ):

                id = f"{customer.email}_{uni.db_name}_{position.id}"
                if not tracking_emails_db.check_if_id_exist(
                    "keep_track_of_sent_emails", id
                ):
                    related_positions.append(
                        [
                            (
                                position.title,
                                position.url,
                                ", ".join(customer_search_result),
                                position.date,
                            ),
                            TrackingEmails(
                                id, uni.db_name, customer.email, position.id
                            ),
                        ]
                    )
        if related_positions:
            log.info(
                f"Sending email to {customer.email} containing {len(related_positions)} positions"
            )
            compose_and_send_email(
                customer.email,
                customer.name,
                uni.name,
                related_positions,
                utils_path,
            )
            log.info(
                "Add information of sended positions to the 'keep_track_of_sent_emails' table"
            )
            for _, email_data in related_positions:
                tracking_emails_db.add_sent_email_data(
                    "keep_track_of_sent_emails", email_data
                )
        else:
            log.info(f"Noting to send to {customer.email} 😿.")

    tracking_emails_cnx.close()
    log.info("Done!")
    log.info("-" * 100)


if __name__ == "__main__":

    for uni in universities:
        filter_and_send_emails(uni)
