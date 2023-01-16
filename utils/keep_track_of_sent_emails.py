import mysql.connector
import os

from dotenv import load_dotenv

from dataclasses import dataclass


@dataclass
class TrackingEmails:
    id: str
    source: str
    customer_email: str
    position_id: int


class TrackingEmailsDatabase:
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

    def create_emails_tracking_table(self, table_name):
        """
        This function creates a table for customer data if it doesn't already exist.

        """
        # Create a cursor object to execute SQL commands
        cursor = self.connection.cursor()
        # id INT AUTO_INCREMENT PRIMARY KEY,

        # Create the table if it doesn't already exist
        sql_query = f"""
        CREATE TABLE IF NOT EXISTS {table_name} (
            id VARCHAR(255) UNIQUE,
            source VARCHAR(255),
            customer_email VARCHAR(255),
            position_id INT(11)
        )
        """

        # Execute the SQL statement

        cursor.execute(sql_query)

        # Commit the changes to the database
        self.connection.commit()

        # Close the cursor
        cursor.close()

    def add_sent_email_data(self, table_name, data):
        """
        This function adds customer data to the specified table.

        Args:
            table_name (str): The name of the table to add data to.
            data (dict): A dictionary containing the customer data to add to the table.
        """
        # Create a cursor object to execute SQL commands
        cursor = self.connection.cursor()

        # Prepare the SQL statement to insert the data into the table
        sql_query = f"REPLACE INTO {table_name} (id, source, customer_email, position_id) VALUES (%s, %s, %s, %s)"
        # Execute the SQL statement
        email_data = (
            data.id,
            data.source,
            data.customer_email,
            data.position_id,
        )
        cursor.execute(sql_query, email_data)

        # Commit the changes to the database
        self.connection.commit()

        # Close the cursor
        cursor.close()

    def get_sent_emails_data(self, table_name):
        """
        This function connects to the MySQL database and retrieves the values from the 'positions' table.
        It returns a list of Position objects, where each object represents a row of the table.
        """
        cursor = self.connection.cursor()

        # SELECT statement to retrieve the values from the positions table
        query = f"SELECT * FROM {table_name}"
        cursor.execute(query)

        # Fetch the rows and create a list of Position objects
        sent_emails = []
        for (id, source, customer_email, position_id) in cursor:
            sent_emails.append(
                TrackingEmails(id, source, customer_email, int(position_id))
            )

        # Close the cursor and connection
        cursor.close()

        return sent_emails

    def check_if_id_exist(self, table_name, id):
        """
        This function connects to the MySQL database and retrieves the values from the 'positions' table.
        It returns a list of Position objects, where each object represents a row of the table.
        """
        cursor = self.connection.cursor()

        # SELECT statement to retrieve the values from the positions table
        query = f"SELECT * FROM {table_name} WHERE id='{id}'"
        cursor.execute(query)

        num_values = len(cursor.fetchall())

        # Close the cursor and connection
        cursor.close()

        return True if num_values >= 1 else False


def main():

    dotenv_path = "/media/hue/Data/codes/fenjan/.env"
    load_dotenv(dotenv_path)
    # Get the database connection details from the .env file
    db_user = os.getenv("DB_USERNAME")
    db_password = os.getenv("DB_PASSWORD")
    db_host = os.getenv("DB_HOST")
    db_port = os.getenv("DB_PORT")
    db_name = os.getenv("DB_NAME")

    db = TrackingEmailsDatabase(
        host=db_host, user=db_user, password=db_password, db=db_name, port=db_port
    )
    print(f"Connecting to {db_name} database.")
    cnx = db.connect_to_database()
    print("Creating 'keep_track_of_sent_emails' Table if not exist.")
    db.create_emails_tracking_table("keep_track_of_sent_emails")
    # source = "tuni_fi"
    # customer_email = "hue.salari@gmail.com"
    # position_id = 21
    # id = f"{customer_email}_{source}_{position_id}"
    # email_data = TrackingEmails(id, source, customer_email, position_id)
    # db.add_sent_email_data(table_name=TRACKING_EMAILS_TABLE_NAME, data=email_data)
    # print(db.get_sent_emails_data(TRACKING_EMAILS_TABLE_NAME))
    cnx.close()


if __name__ == "__main__":
    main()
