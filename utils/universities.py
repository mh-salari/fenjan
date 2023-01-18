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
    "University of Tampere", "tuni_fi", ["Doctoral Researcher"], []
)

Linkoping_University = University(
    "Linköping University", "liu_se", ["PhD"], ["Postdoc"]
)


universities = [
    KTHRoyalInstituteOfTechnology,
    UniversityOfHelsinki,
    UvAUniversityOfAmsterdam,
    UniversityOfTampere,
    Linkoping_University,
]
