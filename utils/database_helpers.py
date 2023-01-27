import os
import mysql.connector
from datetime import datetime

from dotenv import load_dotenv
from dataclasses import dataclass
from utils.customers_database import CustomerDatabase


@dataclass
class Position:
    id: int
    title: str
    url: str
    description: str
    date: str
    scraped_on: datetime


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
        for (id, title, url, descriptions, date, scraped_on) in cursor:
            positions.append(Position(id, title, url, descriptions, date, scraped_on))

        # Close the cursor and connection
        cursor.close()

        return positions


def get_db_connection_values(dotenv_path):

    load_dotenv(dotenv_path)
    # Get the database connection details from the .env file
    db_user = os.getenv("DB_USERNAME")
    db_password = os.getenv("DB_PASSWORD")
    db_host = os.getenv("DB_HOST")
    db_port = os.getenv("DB_PORT")
    db_name = os.getenv("DB_NAME")

    return db_host, db_user, db_password, db_name, db_port


def get_customers_info(dotenv_path):

    customers_db = CustomerDatabase(*get_db_connection_values(dotenv_path))
    customers_cnx = customers_db.connect_to_database()
    customers = customers_db.get_customer_data(table_name="customers")
    customers_cnx.close()

    return customers


def getting_positions_info(dotenv_path, uni):

    positions_db = PositionsDatabase(*get_db_connection_values(dotenv_path))
    positions_cnx = positions_db.connect_to_database()
    positions = positions_db.get_positions(uni.db_name)
    positions_cnx.close()

    return positions
