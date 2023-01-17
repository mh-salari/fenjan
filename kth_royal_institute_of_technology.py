import os
import mysql.connector
from dotenv import load_dotenv
from dataclasses import dataclass

from utils.send_email import send_email
from utils.search import search_for_keywords
from utils.compose_email import compose_email
from utils.customers_database import CustomerDatabase
from utils.keep_track_of_sent_emails import TrackingEmailsDatabase, TrackingEmails

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


@dataclass
class Position:
    id: int
    title: str
    url: str
    description: str
    date: str


class PositionsDatabase:
    def __init__(self, host, user, password, db, port):
        """
        Initializes the PositionsDatabase class

        Args:
            host (str): The hostname or IP address of the MySQL server.
            user (str): The username to connect to the MySQL server.
            password (str): The password to connect to the MySQL server.
            db (str): The name of the database.
        """
        self.host = host
        self.user = user
        self.password = password
        self.db = db
        self.port = port
        self.connection = None

    def connect_to_database(self):
        """
        Connects to the MySQL database and returns a connection object.

        Returns:
            connection : a connection object to the database
        """
        self.connection = mysql.connector.connect(
            host=self.host,
            user=self.user,
            password=self.password,
            database=self.db,
            port=self.port,
        )
        return self.connection

    def get_positions(self, table_name):
        """
        This function connects to the MySQL database and retrieves the values from the 'positions' table.
        It returns a list of Position objects, where each object represents a row of the table.
        """
        cursor = self.connection.cursor()

        # SELECT statement to retrieve the values from the positions table
        query = f"SELECT * FROM {table_name}"
        cursor.execute(query)

        # Fetch the rows and create a list of Position objects
        positions = []
        for (id, title, url, descriptions, date) in cursor:
            positions.append(Position(id, title, url, descriptions, date))

        # Close the cursor and connection
        cursor.close()

        return positions


def compose_and_send_email(recipient_email, recipient_name, positions_list, utils_path):
    """
    Compose and send email containing positions.

    Parameters:
        recipient_email (str): Email address to send the email to.
        recipient_name (str): Customer's name to include in the email.
        positions_list (List[str]): List of positions to include in the email.
        utils_path (str): Base path for the email template file.

    Returns:
        None
    """
    # generate email text using the email template and the given positions

    emails_text = []
    for (title, url, matched_keywords, date), _ in positions_list:
        text = f'<span style="font-weight: bold">Title:</span> {title}\n'
        text += f'\n<span style="font-weight: bold">Matched Keywords:</span>\n<span style="color:#68677b">{matched_keywords}</span>'
        text += f'\n<span style="font-weight: bold">Apply Before:</span> {date}\n'
        text += f"<br><br>🔗: {url}"
        emails_text.append(text)
    email_text = compose_email(
        recipient_name, "KTH Royal Institute of Technology", emails_text, utils_path
    )
    # send email with the generated text
    send_email(
        recipient_email,
        "PhD Positions from KTH Royal Institute of Technology",
        email_text,
        "html",
    )


if __name__ == "__main__":

    # Set base path
    base_path = os.path.dirname(os.path.abspath(__file__))
    utils_path = os.path.join(base_path, "utils")

    dotenv_path = os.path.join(base_path, ".env")
    load_dotenv(dotenv_path)
    # Get the database connection details from the .env file
    db_user = os.getenv("DB_USERNAME")
    db_password = os.getenv("DB_PASSWORD")
    db_host = os.getenv("DB_HOST")
    db_port = os.getenv("DB_PORT")
    db_name = os.getenv("DB_NAME")

    log.info("Getting customers info.")
    customers_db = CustomerDatabase(
        host=db_host, user=db_user, password=db_password, db=db_name, port=db_port
    )
    customers_cnx = customers_db.connect_to_database()
    customers = customers_db.get_customer_data(table_name="customers")
    customers_cnx.close()

    log.info("Getting KTH Royal Institute of Technology positions info.")
    positions_db = PositionsDatabase(
        host=db_host, user=db_user, password=db_password, db=db_name, port=db_port
    )
    positions_cnx = positions_db.connect_to_database()
    positions = positions_db.get_positions("kth_se")
    positions_cnx.close()

    log.info("Getting sent_emails info.")
    tracking_emails_db = TrackingEmailsDatabase(
        host=db_host, user=db_user, password=db_password, db=db_name, port=db_port
    )

    tracking_emails_cnx = tracking_emails_db.connect_to_database()

    sent_emails = tracking_emails_db.get_sent_emails_data("keep_track_of_sent_emails")

    for customer in customers:
        customer_keywords = list(
            set(
                [keyword.replace(" ", "") for keyword in customer.keywords]
                + customer.keywords
            )
        )

        log.info("Filtering Positions based of customer keywords")
        related_positions = []
        for position in positions:
            search_result = search_for_keywords(
                position.title + position.description, keywords=customer_keywords
            )
            if search_result:
                id = f"{customer.email}_{'kth_se'}_{position.id}"
                if not tracking_emails_db.check_if_id_exist(
                    "keep_track_of_sent_emails", id
                ):
                    related_positions.append(
                        [
                            (
                                position.title,
                                position.url,
                                ", ".join(search_result),
                                position.date,
                            ),
                            TrackingEmails(id, "kth_se", customer.email, position.id),
                        ]
                    )
        if related_positions:
            log.info(
                f"Sending email to {customer.email} containing {len(related_positions)} positions"
            )
            compose_and_send_email(
                customer.email, customer.name, related_positions, utils_path
            )
            log.info(
                "Add information of sended positions to the 'keep_track_of_sent_emails' table"
            )
            for _, email_data in related_positions:
                tracking_emails_db.add_sent_email_data(
                    "keep_track_of_sent_emails", email_data
                )
    tracking_emails_cnx.close()
    log.info("Done!")
