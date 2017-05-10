package pg2mysql_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/pg2mysql"
)

var _ = Describe("Verifier", func() {
	var (
		verifier pg2mysql.Verifier
		mysql    pg2mysql.DB
		pg       pg2mysql.DB
	)

	BeforeEach(func() {
		mysql = pg2mysql.NewMySQLDB(
			mysqlRunner.DBName,
			"root",
			"",
			"127.0.0.1",
			3306,
		)

		err := mysql.Open()
		Expect(err).NotTo(HaveOccurred())

		pg = pg2mysql.NewPostgreSQLDB(
			pgRunner.DBName,
			"",
			"",
			"127.0.0.1",
			5432,
			"disable",
		)
		err = pg.Open()
		Expect(err).NotTo(HaveOccurred())

		verifier = pg2mysql.NewVerifier(pg, mysql)
	})

	AfterEach(func() {
		err := mysql.Close()
		Expect(err).NotTo(HaveOccurred())
		err = pg.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Verify", func() {
		It("returns results", func() {
			result, err := verifier.Verify()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(2))
			Expect(result).To(ContainElement(pg2mysql.VerificationResult{
				TableName:       "table_with_id",
				MissingRowCount: 0,
			}))

			Expect(result).To(ContainElement(pg2mysql.VerificationResult{
				TableName:       "table_without_id",
				MissingRowCount: 0,
			}))
		})

		Context("when there is data in postgres that is not in mysql", func() {
			BeforeEach(func() {
				result, err := pgRunner.DB().Exec("INSERT INTO table_with_id (id, name, ci_name, created_at, truthiness) VALUES (3, 'some-name', 'some-ci-name', now(), false);")
				Expect(err).NotTo(HaveOccurred())
				rowsAffected, err := result.RowsAffected()
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsAffected).To(BeNumerically("==", 1))
			})

			It("returns a result", func() {
				result, err := verifier.Verify()
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(2))
				Expect(result).To(ContainElement(pg2mysql.VerificationResult{
					TableName:       "table_with_id",
					MissingRowCount: 1,
				}))

				Expect(result).To(ContainElement(pg2mysql.VerificationResult{
					TableName:       "table_without_id",
					MissingRowCount: 0,
				}))
			})
		})

		Context("when there is data in postgres that is in mysql", func() {
			BeforeEach(func() {
				id := 3
				name := "some-name"
				ciname := "some-ci-name"
				created_at := time.Now().UTC()
				truthiness := true

				stmt := "INSERT INTO table_with_id (id, name, ci_name, created_at, truthiness) VALUES ($1, $2, $3, $4, $5);"
				result, err := pgRunner.DB().Exec(stmt, id, name, ciname, created_at, truthiness)
				Expect(err).NotTo(HaveOccurred())
				rowsAffected, err := result.RowsAffected()
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsAffected).To(BeNumerically("==", 1))

				stmt = "INSERT INTO table_with_id (id, name, ci_name, created_at, truthiness) VALUES (?, ?, ?, ?, ?);"
				result, err = mysqlRunner.DB().Exec(stmt, id, name, ciname, created_at, truthiness)
				Expect(err).NotTo(HaveOccurred())
				rowsAffected, err = result.RowsAffected()
				Expect(err).NotTo(HaveOccurred())
				Expect(rowsAffected).To(BeNumerically("==", 1))
			})

			It("returns a result", func() {
				result, err := verifier.Verify()
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(2))
				Expect(result).To(ContainElement(pg2mysql.VerificationResult{
					TableName:       "table_with_id",
					MissingRowCount: 0,
				}))

				Expect(result).To(ContainElement(pg2mysql.VerificationResult{
					TableName:       "table_without_id",
					MissingRowCount: 0,
				}))
			})
		})
	})
})
