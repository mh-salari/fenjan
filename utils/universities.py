class University:
    def __init__(self, name, db_name, target_keywords=[], forbidden_keywords=[]):
        self.name = name
        self.db_name = db_name
        self.target_keywords = target_keywords
        self.forbidden_keywords = forbidden_keywords


KTHRoyalInstituteOfTechnology = University(
    "KTH Royal Institute of Technology", "kth_se", [], []
)

UniversityOfHelsinki = University(
    "University of Helsinki", "helsinki_fi", ["Doctoral Researcher"], ["Postdoctoral"]
)

UvAUniversityOfAmsterdam = University("UvA University of Amsterdam", "uva_nl", [], [])

UniversityOfTampere = University(
    "University of Tampere", "tuni_fi", ["Doctoral Researcher"], ["Postdoctoral"]
)

Linkoping_University = University(
    "Linköping University", "liu_se", ["PhD"], ["Postdoc"]
)

TechnicalUniversityOfMunich = University(
    "Technical University of Munich (TUM)",
    "tum_de",
    ["Doctoral", "PhD", "Ph.D."],
    ["Postdoctoral"],
)

FreieUniversitatBerlin = University(
    "Freie Universität Berlin", "fu_berlin_de", ["PhD", "Ph.D."], []
)

KarlsruheInstituteOfTechnology = University(
    "Karlsruhe Institute of Technology (KIT)", "kit_edu", [], []
)

UniversityOfTurku = University(
    "University of Turku", "utu_fi", ["Doctoral Researcher"], ["Postdoctoral"]
)

LappeenrantaUniversityOfTechnology = University(
    "Lappeenranta University of Technology",
    "lut_fi",
    ["Junior researcher", "PhD", "Ph.D.", "doctoral candidate"],
    [],
)
UniversityOfOulu = University(
    "University of Oulu",
    "oulu_fi",
    ["Doctoral Researcher", "PhD", "Ph.D."],
    ["Postdoctoral"],
)


UniversityOfEasternFinland = University(
    "University of Eastern Finland",
    "uef_fi",
    ["Doctoral Researcher", "PhD", "Ph.D."],
    ["Postdoctoral"],
)

ChalmersUniversityOfTechnology = University(
    "Chalmers University of Technology", "chalmers_se", ["PhD"], []
)

AaltoUniversity = University("Aalto University", "aalto_fi", [], [])

LundUniversity = University(
    "Lund University",
    "lunduniversity_lu_se",
    ["Doctoral Student", "PhD", "Ph.D."],
    ["Postdoctoral"],
)

UppsalaUniversity = University("UppsalaUniversity", "uu_se", [], [])

universities = [
    KTHRoyalInstituteOfTechnology,
    UniversityOfHelsinki,
    UvAUniversityOfAmsterdam,
    UniversityOfTampere,
    Linkoping_University,
    TechnicalUniversityOfMunich,
    FreieUniversitatBerlin,
    KarlsruheInstituteOfTechnology,
    UniversityOfTurku,
    LappeenrantaUniversityOfTechnology,
    UniversityOfOulu,
    UniversityOfEasternFinland,
    ChalmersUniversityOfTechnology,
    AaltoUniversity,
    LundUniversity,
    UppsalaUniversity,
]
